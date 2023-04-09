package main

import (
	"flag"
	"os"
	"net"
	"net/http"
	"io"
	"log"
	"fmt"
	"time"
	"encoding/json"
	"pikpak-upload-server/model"
	"golang.org/x/net/websocket"
)

var PATH = "./"

func main() {
	log_file, _ := os.OpenFile(PATH+"pikpak_log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0644)
	log.SetOutput(log_file)
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #1!\n")
	}
	http.HandleFunc("/", h1)
	http.HandleFunc("/api", h2)
	http.HandleFunc("/log", h3)
	http.Handle("/ssh", websocket.Handler(sshServer))
	PORT := ":" + os.Getenv("PORT")
	fmt.Println(PORT)
	log.Println("listen port: ", PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))
}

func h2(w http.ResponseWriter, r *http.Request) {
	type Args struct {
		Fn   string `json:"fn"`
		Link string `json:"link"`
	}
	var args Args
	rbody, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(rbody[0:], &args)
	go job(args.Fn, args.Link)
	io.WriteString(w, "good lucky\n")
}

func job(fn, link string) {
	//download file
	defer os.Remove(PATH+fn)
	f, err := os.OpenFile(PATH+fn, os.O_RDWR|os.O_CREATE, 0755)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return
	}
	var buf [32 * 1024 * 1024]byte
	log.Println("download start ...")
	res, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		n, err := res.Body.Read(buf[0:])
		if err != nil && err != io.EOF {
			log.Println(err)
			return
		}
		//write to file
		_, f_err := f.Write(buf[0:n])
		if f_err != nil {
			log.Println(f_err)
			return
		}
		if err == io.EOF {
			break
		}
	}
	t := fmt.Sprintf("%d",time.Now().Unix())
	f.Write([]byte(t))
	f.Close()
	log.Println("download completed!")
	//upload file
	p := model.NewPikPak("gungfu2022@hotmail.com", "***")
	err = p.Login()
	if err != nil {
		log.Println(err)
	}
	err = p.AuthCaptchaToken("POST:/drive/v1/files")
	if err != nil {
		log.Println(err)
	}
	log.Println("upload start...")
	err = p.UploadFile(parentId, PATH+fn)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("upload completed!")
	}
}

var parentId = ""

func init() {
	parentid := flag.String("p", "", "ParentId")
	concurrent := flag.Int("c", 4, "Concurrent")
	flag.Parse()
	parentId = *parentid
	model.Concurrent = int64(*concurrent)
}

func h3(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(PATH+"pikpak_log")
	if err != nil {
		log.Println(err)
		io.WriteString(w, err.Error())
	}
	w.Write(data)
}

func sshServer(ws *websocket.Conn) {
	conn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}
	log.Println("connect local ssh server success")
	go io.Copy(ws, conn)
	io.Copy(conn, ws)
}
