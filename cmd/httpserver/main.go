package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/jmservic/httpfromtcp/internal/headers"
	"github.com/jmservic/httpfromtcp/internal/request"
	"github.com/jmservic/httpfromtcp/internal/response"
	"github.com/jmservic/httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, proxyHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func mainHandler(w *response.Writer, req *request.Request) {

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		badRequest(w)
	case "/myproblem":
		internalError(w)
	case "/video":
		videoResponse(w)
	default:
		goodRequest(w)
	}
}

func proxyHandler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		url := "https://httpbin.org/" + path
		fmt.Println("Proxing to", url)
		resp, err := http.Get(url)
		if err != nil {
			badRequest(w)
		}

		h := response.GetDefaultHeaders(0)
		//for key, vals := range resp.Header {
		//	for _, val := range vals {
		//		headers.Set(key, val)
		//	}
		//}
		h.Delete("Content-Length")
		//h.Delete("Connection")
		h.Replace("Transfer-Encoding", "chunked")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		w.WriteStatusLine(response.StatusCode(resp.StatusCode))
		w.WriteHeaders(h)

		const maxChunkSize = 1024
		fullResponse := make([]byte, 0)
		buffer := make([]byte, maxChunkSize)

		for {
			n, err := resp.Body.Read(buffer)
			fmt.Printf("Read %d bytes from the response body\n", n)
			if n > 0 {
				fullResponse = append(fullResponse, buffer[:n]...)
				_, err = w.WriteChunkedBody(buffer[:n])
				if err != nil {
					fmt.Println("Error writing chunked body:")
					break
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading response body:", err)
				break
			}
		}

		_, err = w.WriteChunkedBodyDone(true)
		if err != nil {
			fmt.Println(err)
		}
		trailers := headers.NewHeaders()
		responseHash := sha256.Sum256(fullResponse)
		//	fmt.Printf("%x | %d\n", responseHash, len(fullResponse))
		//	fmt.Print(string(fullResponse))
		//trailers["X-Content-Sha256"] = fmt.Sprintf("%x", responseHash)
		//trailers["X-Content-Length"] = fmt.Sprintf("%d", len(fullResponse))
		trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", responseHash))
		trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullResponse)))
		//	hResponse, _ := trailers.Get("X-Content-SHA256")
		//	hLength, _ := trailers.Get("X-Content-Length")
		//	fmt.Println(hResponse, hLength)
		err = w.WriteTrailers(trailers)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		mainHandler(w, req)
	}
}

func badRequest(w *response.Writer) {
	body := []byte(`<html>
	  <head>
	    <title>400 Bad Request</title>
	  </head>
	  <body>
	    <h1>Bad Request</h1>
	    <p>Your request honestly kinda sucked.</p>
	  </body>
	</html>`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Replace("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusBadRequest)
	w.WriteHeaders(headers)
	w.WriteBody(body)
}

func internalError(w *response.Writer) {
	body := []byte(`<html>
	  <head>
	    <title>500 Internal Server Error</title>
	  </head>
	  <body>
	    <h1>Internal Server Error</h1>
	    <p>Okay, you know what? This one is on me.</p>
	  </body>
	</html>`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Replace("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusInternalServerError)
	w.WriteHeaders(headers)
	w.WriteBody(body)
}

func goodRequest(w *response.Writer) {
	body := []byte(`<html>
	  <head>
	    <title>200 OK</title>
	  </head>
	  <body>
	    <h1>Success!</h1>
	    <p>Your request was an absolute banger.</p>
	  </body>
	</html>`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Replace("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusInternalServerError)
	w.WriteHeaders(headers)
	w.WriteBody(body)
}

func videoResponse(w *response.Writer) {
	video, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		fmt.Printf("error opening the video file: %v\n", err)
		return
	}
	headers := response.GetDefaultHeaders(len(video))
	headers.Replace("Content-Type", "video/mp4")
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(headers)
	w.WriteBody(video)
}
