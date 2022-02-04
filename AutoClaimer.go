package main

import (
	"bufio"
	"crypto/tls"
	_ "crypto/tls"
	_ "encoding/base64"
	_ "fmt"
	"github.com/DisgoOrg/disgohook"
	"github.com/DisgoOrg/disgohook/api"
	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	_ "strings"
	"sync"
	"time"
)

// placeholder url. Change it to your base url.
const BASE_URL = "https://i.instagram.com/api/v1/accounts/set_username/"
var c *fasthttp.Client
// number of parallelism
var writer = uilive.New()
var counter = 0
var GoodAttempts = 0
var BadAttempts = 0
var Claimed string
var proxies []string
var proxiescount int = 1
var users []string
var userscount int = 1
var sessions []string
var sessionscount int = 1
var rateCount = 10
var threadsCount = 50
// Stream inputs to input channel


func sendDiscord(username string) {
	webhook, _ := disgohook.NewWebhookClientByToken(nil, nil, "Discord Token Here")
	_, _ = webhook.SendEmbeds(api.NewEmbedBuilder().
		AddField("Username",username,true).
		AddField("R/N", strconv.Itoa(counter),true).
		SetAuthor("New Claim","","https://cdn.discordapp.com/avatars/225147959757635584/a_55374353240a1adca3d749c19bd1883a.gif?size=1024").
		SetFooter("Coded With Luv , L0N3LY","").
		SetThumbnail("https://cdn.discordapp.com/avatars/225147959757635584/a_55374353240a1adca3d749c19bd1883a.gif?size=1024").
		Build(),
	)

	return
}



// Normal function for HTTP call, no knowledge of goroutine/channels
func sendUser(username,proxy,session string) (string, error) {
	c = &fasthttp.Client{
		Dial: fasthttpproxy.FasthttpHTTPDialer(proxy),
		MaxConnsPerHost:     1000000,
		ReadBufferSize:      8192000,
		ReadTimeout:         time.Duration(1) * time.Second,
		MaxIdleConnDuration: time.Duration(1) * time.Second,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var jsonstring = []byte("username="+username)
	var Body string
	req := fasthttp.AcquireRequest()
	req.SetBody(jsonstring)
	req.Header.SetMethod("POST")
	req.Header.Add("User-Agent", "Instagram 167.0.0.24.120 Android (25/7.1.2; 192dpi; 720x1280; google; G011A; G011A; intel; en_US; 256966583)")
	req.Header.Add("Cookie", "sessionid="+session)
	req.Header.Add("X-MID", "missing")
	req.Header.SetContentType("application/x-www-form-urlencoded; charset=UTF-8")
	req.SetRequestURI(BASE_URL)
	res := fasthttp.AcquireResponse()
	err := c.Do(req, res)
	if err != nil {
		//fmt.Println("h1")
		counter += 1
		BadAttempts+=1
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
		if proxiescount == len(proxies)-1 || proxiescount > len(proxies) {
			proxiescount = 1
		} else if proxiescount < len(proxies)-1 {
			proxiescount += 1
		}
		if sessionscount == len(sessions)-1 || sessionscount > len(sessions) {
			sessionscount = 1
		} else if sessionscount < len(sessions) {
			sessionscount += 1
		}
		_, _ = sendUser(username, proxies[proxiescount], sessions[sessionscount])
		return "",nil
	} else {
		Body  = string(res.Body())
		//fmt.Println("h2")
		if strings.Contains(Body, "\""+username+"\"") {
			if strings.Contains(Claimed,username) {

			}else{
				if Claimed == "" {
					Claimed = username
				} else {
					Claimed = Claimed + "," + username
				}
				f, err := os.Create(username + ".txt")
				if err != nil {
					log.Fatal(err)
				}
				defer func(f *os.File) {
					err := f.Close()
					if err != nil {

					}
				}(f)
				_, err2 := f.WriteString(username + ":" + session + "\n")
				if err2 != nil {
					log.Fatal(err2)
				}
				sendDiscord(username)
			}
		} else if Body == "" {
			//fmt.Println("h3")
			counter+=1
			BadAttempts+=1
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(res)
			if proxiescount == len(proxies)-1 || proxiescount > len(proxies) {
				proxiescount = 1
			} else if proxiescount < len(proxies)-1 {
				proxiescount += 1
			}
			if sessionscount == len(sessions)-1 || sessionscount > len(sessions) {
				sessionscount = 1
			} else if sessionscount < len(sessions) {
				sessionscount += 1
			}
			_, _ = sendUser(username, proxies[proxiescount], sessions[sessionscount])
			return "", nil
		} else {
			//fmt.Println("h4")
			counter += 1
			GoodAttempts+=1
		}
	}
	jsonstring = nil
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return Body, nil
}

// Wrapper for sendUser return value, used as result channel type


func AsyncHTTP(users []string) ([]string, error) {
	// limit concurrency to 5

	// have a max rate of 10/sec
	rate := make(chan struct{}, rateCount)
	for i := 0; i < cap(rate); i++ {
		rate <- struct{}{}
	}

	// leaky bucket

	var wg sync.WaitGroup
	for i := 0; i <= threadsCount; i++ {
		wg.Add(1)
		go func() {

			defer wg.Done()
			// wait for the rate limiter

			if userscount == len(users)-1 || userscount > len(users) {
				userscount = 1
			} else if userscount < len(users)-1 {
				userscount += 1
			}
			if proxiescount == len(proxies)-1 || proxiescount > len(proxies) {
				proxiescount = 1
			} else if proxiescount < len(proxies)-1 {
				proxiescount += 1
			}
			if sessionscount == len(sessions)-1 || sessionscount > len(sessions) {
				sessionscount = 1
			} else if sessionscount < len(sessions) {
				sessionscount += 1
			}
			_, _ = sendUser(users[userscount], proxies[proxiescount], sessions[sessionscount])
			d := color.New(color.FgCyan, color.Bold)
			_, _ = d.Fprintf(writer, "[1] - Attempts := {%v}\n",counter)
			_, _ = d.Fprintf(writer.Newline(), "[2] - Good     := {%v}\n",GoodAttempts)
			_, _ = d.Fprintf(writer.Newline(), "[3] - Error    := {%v}\n",BadAttempts)
			_, _ = d.Fprintf(writer.Newline(), "[4] - User     := {%v}\n",users[userscount])
			_, _ = d.Fprintf(writer.Newline(), "[5] - Proxy    := {%v}\n",proxies[proxiescount])
			_, _ = d.Fprintf(writer.Newline(), "[6] - Session  := {%v}\n",sessions[sessionscount])
			_, _ = d.Fprintf(writer.Newline(), "[7] - Claimed  := {%v}\n",Claimed)
		}()
	}

	wg.Wait()
	close(rate)

	return users, nil
}

func main() {
	//resp, err := http.Get("https://pastebin.com/raw/gzDHSvN6")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//We Read the response body on the line below.
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//Convert the body to type string
	//sb := string(body)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	//ipadd := conn.LocalAddr().(*net.UDPAddr)
	//if strings.Contains(sb,ipadd.IP.String()){
		scannert := bufio.NewScanner(os.Stdin)
		d := color.New(color.BgHiRed, color.Bold)
		d.Println("[1] - Enter Threads :")
		scannert.Scan()
		threadsCount, _ = strconv.Atoi(scannert.Text())
		d.Println("[2] - Enter Rate :")
		scannert.Scan()
		rateCount, _ = strconv.Atoi(scannert.Text())
		CallClear()
		// populate users param
		file, err := os.Open("list.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err = file.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {             // internally, it advances token based on sperator
			users = append(users, scanner.Text())
		}

		file2, err2 := os.Open("proxy.txt")
		if err2 != nil {
			log.Fatal(err2)
		}
		defer func() {
			if err2 = file2.Close(); err2 != nil {
				log.Fatal(err2)
			}
		}()

		scanner2 := bufio.NewScanner(file2)

		for scanner2.Scan() {             // internally, it advances token based on sperator
			proxies = append(proxies, scanner2.Text())
		}

		file3, err3 := os.Open("session.txt")
		if err3 != nil {
			log.Fatal(err3)
		}
		defer func() {
			if err3 = file3.Close(); err3 != nil {
				log.Fatal(err3)
			}
		}()

		scanner3 := bufio.NewScanner(file3)

		for scanner3.Scan() {             // internally, it advances token based on sperator
			sessions = append(sessions, scanner3.Text())
		}
		color.Cyan("█████████████████████████████████████\n█▄─▄███─▄▄─█▄─▀█▄─▄█▄▄▄░█▄─▄███▄─█─▄█\n██─██▀█─██─██─█▄▀─███▄▄░██─██▀██▄─▄██\n▀▄▄▄▄▄▀▄▄▄▄▀▄▄▄▀▀▄▄▀▄▄▄▄▀▄▄▄▄▄▀▀▄▄▄▀▀")
		writer.Start()
			 for{
				_, err := AsyncHTTP(users)
				if err != nil {
					return
				}
			}


	//}else{
	//	fmt.Println("Not Activated ...")
	//}

}
var clear map[string]func()
func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok { //if we defined a clear func for that platform:
		value()  //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}