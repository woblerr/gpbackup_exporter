package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		t.Errorf("\nGet error during generate random int value:\n%v", err)
	}
	port := 50000 + int(n.Int64())
	os.Args = []string{"gpbackup_exporter", "--prom.port=" + strconv.Itoa(port)}
	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()
	time.Sleep(time.Second)
	urlList := []string{
		fmt.Sprintf("http://localhost:%d/metrics", port),
		fmt.Sprintf("http://localhost:%d/", port),
		fmt.Sprintf("http://localhost:%d/health", port),
	}
	for _, url := range urlList {
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("\nGet error during GET:\n%v", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("\nGet bad response code:\n%v\nwant:\n%v", resp.StatusCode, 200)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("\nGet error during read resp body:\n%v", err)
		}
		if len(string(b)) == 0 {
			t.Errorf("\nGet zero body:\n%s", string(b))
		}
	}
}
