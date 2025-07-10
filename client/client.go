package client

import (
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	cl       *http.Client
	mkDaemon func() error
}

func New(baseURL string, mkDaemon func() error) *Client {

	return &Client{baseURL: baseURL, cl: &http.Client{}, mkDaemon: mkDaemon}
}

func (c *Client) ensureDaemonAlive() error {

	_, err := c.cl.Get(c.baseURL + "/up")
	if err == nil {
		return nil
	}
	err = c.mkDaemon()
	if err != nil {
		return fmt.Errorf("failed to start daemon: %s", err)
	}

	for range 3 {
		_, err = c.cl.Get(c.baseURL + "/up")
		if err != nil {
			err = fmt.Errorf("daemon still not alive after startup: %s", err)
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (c *Client) StartActivity(name string, ignoreIdle bool) error {

	err := c.ensureDaemonAlive()
	if err != nil {
		return err
	}
	url := c.baseURL + "/start"
	if ignoreIdle {
		url = url + "?ignoreIdle=true"
	}
	post, err := c.cl.Post(url, "text/plain", strings.NewReader(name))
	if err != nil {
		return err
	}
	defer post.Body.Close()
	b, err := io.ReadAll(post.Body)
	if err != nil {
		return err
	}
	if post.StatusCode != 200 {
		return fmt.Errorf("unexpected status code in StartActivity: %d, %s", post.StatusCode, b)
	}
	return nil
}

func (c *Client) StopActivity() error {

	err := c.ensureDaemonAlive()
	if err != nil {
		return err
	}
	post, err := c.cl.Post(c.baseURL+"/end", "text/plain", nil)
	if err != nil {
		return err
	}
	defer post.Body.Close()
	b, err := io.ReadAll(post.Body)
	if err != nil {
		return err
	}
	if post.StatusCode != 200 {
		return fmt.Errorf("unexpected status code in StopActivity: %d, %s", post.StatusCode, b)
	}
	return nil
}

func (c *Client) TerminateDaemon() error {

	post, err := c.cl.Post(c.baseURL+"/terminate", "text/plain", nil)
	if err != nil {
		return err
	}
	defer post.Body.Close()
	b, err := io.ReadAll(post.Body)
	if err != nil {
		return err
	}
	if post.StatusCode != 200 {
		return fmt.Errorf("unexpected status code in TerminateDaemon: %d, %s", post.StatusCode, b)
	}
	return nil
}

func (c *Client) GetStatus() (string, error) {

	err := c.ensureDaemonAlive()
	if err != nil {
		return "", err
	}
	resp, err := c.cl.Get(c.baseURL + "/status")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code in GetStatus: %d, %s", resp.StatusCode, b)
	}
	return string(b), nil
}

func (c *Client) Aggregate(start, end time.Time, round bool) (string, error) {

	err := c.ensureDaemonAlive()
	if err != nil {
		return "", err
	}
	url := c.baseURL + "/aggregate"
	q := make(nurl.Values)
	if !start.IsZero() {
		q.Set("start", start.Format(time.DateOnly))
	}
	if !end.IsZero() {
		q.Set("end", end.Format(time.DateOnly))
	}
	q.Set("round", strconv.FormatBool(round))
	qs := q.Encode()
	if qs != "" {
		url = url + "?" + qs
	}
	resp, err := c.cl.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code in Aggregate: %d, %s", resp.StatusCode, b)
	}
	return string(b), nil
}
