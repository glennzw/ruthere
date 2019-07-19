/* ruthere - Are You There?
*
* A tiny endpoint to query the response code of a supplied URL via a returned images properties.
*
* If you don't know why this is useful, you probably don't need it.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	badURL             = 50
	badConnect         = 40
	serviceUnavailable = 30
)

type message struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func landing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, friend. Example usage: %s", (r.Host + "/u/www.google.com"))
}

func buildImage(status int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, status, status))
	cyan := color.RGBA{100, 200, 200, 0xff}
	draw.Draw(img, img.Bounds(), &image.Uniform{cyan}, image.ZP, draw.Src)
	buffer := new(bytes.Buffer)
	jpeg.Encode(buffer, img, nil)
	return buffer.Bytes()
}

func fetchURLJS(w http.ResponseWriter, r *http.Request) {

	defer func() { //Handle panic.
		if recover() != nil {
			log.Println("[!] Panic! Returning error image.")
			imageBytes := buildImage(serviceUnavailable)
			w.Write(imageBytes)
		}
	}()

	status := 1
	iurl := r.URL.String()[3:] //Hack to allow ? in URL
	remoteIP := r.RemoteAddr
	remoteIP, _, err := net.SplitHostPort(remoteIP)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	//URL rejiggering. Mux doesn't like forward slashes, sometimes. http:// gets converted to http:/ but https:// remains intact, usually.
	error := ""
	if len(iurl) < 5 || iurl[0:4] != "http" {
		iurl = "http://" + iurl
	}
	if len(iurl) < 8 {
		log.Printf("[!] [%s] Error parsing URL: %s", remoteIP, iurl)
		status = badURL
	} else {

		if iurl[0:6] == "http:/" && iurl[6:7] != "/" {
			iurl = "http://" + iurl[6:]
		} else if iurl[0:7] == "https:/" && iurl[7:8] != "/" {
			iurl = "https://" + iurl[7:]
		}
		_, err := url.ParseRequestURI(iurl)
		if err != nil {
			log.Printf("[!] [%s] Error parsing URL: %s", remoteIP, iurl)
			status = badURL
			error = err.Error()
		} else {
			var netClient = &http.Client{
				Timeout: time.Second * 10,
			}
			response, err := netClient.Get(iurl)
			if err != nil {
				log.Printf("[!] [%s] Error connecting to URL: %s (%s)", remoteIP, iurl, err)
				status = badConnect
				error = err.Error()
			} else {
				status = response.StatusCode
				log.Printf("[+] [%s] Returning %s status for %s\n", remoteIP, strconv.Itoa(status), iurl)
				defer response.Body.Close()
			}
		}
	}
	//Return status code as JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("User-Agent", "ruthere")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(message{status, error})

}

func fetchURL(w http.ResponseWriter, r *http.Request) {

	defer func() { //Handle panic.
		if recover() != nil {
			log.Println("[!] Panic! Returning error image.")
			imageBytes := buildImage(serviceUnavailable)
			w.Write(imageBytes)
		}
	}()

	status := 1
	//vars := mux.Vars(r)
	//iurl := vars["url"]
	iurl := r.URL.String()[3:] //Hack to allow ? in URL
	remoteIP := r.RemoteAddr
	remoteIP, _, err := net.SplitHostPort(remoteIP)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	//URL rejiggering. Mux doesn't like forward slashes, sometimes. http:// gets converted to http:/ but https:// remains intact, usually.
	if len(iurl) < 5 || iurl[0:4] != "http" {
		iurl = "http://" + iurl
	}
	if len(iurl) < 8 {
		log.Printf("[!] [%s] Error parsing URL: %s", remoteIP, iurl)
		status = badURL
	} else {

		if iurl[0:6] == "http:/" && iurl[6:7] != "/" {
			iurl = "http://" + iurl[6:]
		} else if iurl[0:7] == "https:/" && iurl[7:8] != "/" {
			iurl = "https://" + iurl[7:]
		}
		_, err := url.ParseRequestURI(iurl)
		if err != nil {
			log.Printf("[!] [%s] Error parsing URL: %s", remoteIP, iurl)
			status = badURL
		} else {
			var netClient = &http.Client{
				Timeout: time.Second * 10,
			}
			response, err := netClient.Get(iurl)
			if err != nil {
				log.Printf("[!] [%s] Error connecting to URL: %s (%s)", remoteIP, iurl, err)
				status = badConnect
			} else {
				status = response.StatusCode
				log.Printf("[+] [%s] Returning %d x %d image for %s\n", remoteIP, status, status, iurl)
				defer response.Body.Close()
			}
		}
	}
	//Return image as status code
	imageBytes := buildImage(status)
	w.Write(imageBytes)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("No Gorilla!\n"))
}

func main() {
	port := os.Getenv("PORT")

	r := mux.NewRouter()
	r.SkipClean(true) //Mux doesn't like double forward slashes
	r.HandleFunc("/", landing).Methods("GET")
	r.HandleFunc("/u/{url:[a-z,A-Z,0-9,\\-,_,\\/,\\.,:,\\/\\/]*}", fetchURL).Methods("GET")   //Mux doesn't like forward slashes
	r.HandleFunc("/j/{url:[a-z,A-Z,0-9,\\-,_,\\/,\\.,:,\\/\\/]*}", fetchURLJS).Methods("GET") //j endpoint for JavaScript version (thanks @KentonVarda)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	http.ListenAndServe(":"+port, r)
}
