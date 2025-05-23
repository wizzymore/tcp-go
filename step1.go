package main

import (
	"encoding/json"
	"io"
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StepOneRequest struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type StepOneResponse struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func handle_step_one(log zerolog.Logger, conn net.Conn, request []byte) bool {
	log.Debug().Str("request", string(request)).Msg("Got message from client")
	var requestData StepOneRequest
	if err := json.Unmarshal(request, &requestData); err != nil {
		log.Error().Err(err).Msg("Could not parse JSON request")
		return false
	}
	if requestData.Method != "isPrime" || requestData.Number == nil {
		return false
	}
	var responseData StepOneResponse
	responseData.Method = "isPrime"
	if *requestData.Number == float64(int(*requestData.Number)) {
		responseData.Prime = isPrime(int(*requestData.Number))
	} else {
		responseData.Prime = false
	}
	data, _ := json.Marshal(responseData)
	data = append(data, '\n')
	conn.Write(data)
	return true
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func step_one(conn net.Conn) {
	log := log.With().Str("remote_addr", conn.RemoteAddr().String()).Logger()
	defer func() {
		log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("Closing connection")
		conn.Close()
	}()

	left := []byte{}
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		log.Info().Msgf("Got new buffer %s", strings.ReplaceAll(string(buf), "\n", "\\n"))
		if err != nil {
			if err == io.EOF {
				log.Info().Msg("Client closed the connection")
				return
			}
			log.Err(err).Msg("Client read error")
			return
		}

		start := 0
		for i := range buf[:n] {
			if buf[i] == '\n' {
				left = append(left, buf[start:i]...)
				if !handle_step_one(log, conn, left) {
					conn.Write([]byte("bye, bye\n"))
					return
				}
				left = []byte{}
				start = i + 1
			} else if i == n-1 {
				left = append(left, buf[start:n]...)
			}
		}
	}
}
