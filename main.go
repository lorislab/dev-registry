package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

var manifestRegexep = regexp.MustCompile(`^\/v2\/(.+)\/manifests\/(.+)$`)

func main() {
	log.SetPrefix("[PROXY] ")

	registyHost := env("PROXY_DOCKER_REGISTRY_HOST", "127.0.0.1") + ":" + env("PROXY_DOCKER_REGISTRY_PORT", "5002")
	handler := &proxy{
		dockerRegistryUrl: env("PROXY_DOCKER_REGISTRY_PROTOCOL", "http://") + registyHost,
		dockerRegistry:    registyHost,
	}
	log.Printf("Proxy for docker registry %s / %s started.", handler.dockerRegistry, handler.dockerRegistryUrl)
	log.Print("ListenAndServe:", http.ListenAndServe(env("PROXY_LISTENER", ":5000"), handler))
}

type proxy struct {
	dockerRegistryUrl string
	dockerRegistry    string
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	if req.Method == "GET" {
		matches := manifestRegexep.FindStringSubmatch(req.URL.Path)
		if len(matches) == 3 {
			if err := updateDockerInRegistry(p.dockerRegistry, matches[1], matches[2]); err != nil {
				handleError(wr, err)
				return
			}
		}
	}
	client := &http.Client{}

	u, e := url.Parse(p.dockerRegistryUrl + req.URL.String())
	if e != nil {
		handleError(wr, e)
		return
	}
	req.URL = u

	req.RequestURI = ""
	delHopHeaders(req.Header)

	// execute client
	resp, err := client.Do(req)
	if err != nil {
		handleError(wr, err)
		return
	}
	defer resp.Body.Close()

	// copy client response
	delHopHeaders(resp.Header)
	copyHeader(wr.Header(), resp.Header)

	wr.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(wr, resp.Body); err != nil {
		handleError(wr, err)
		return
	}
}

func handleError(wr http.ResponseWriter, e error) {
	if e != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Print("Error ServeHTTP:", e)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func env(name string, defaultValue string) string {
	if v, e := os.LookupEnv(name); e {
		return v
	}
	return defaultValue
}

func updateDockerInRegistry(registry, image, tag string) error {
	name := image + ":" + tag
	registryName := registry + "/" + name

	if err := cmd("docker", "tag", name, registryName); err != nil {
		return err
	}
	if err := cmd("docker", "push", registryName); err != nil {
		return err
	}
	return nil
}

func cmd(name string, params ...string) error {
	log.Printf("%s %s", name, params)
	cmd := exec.Command(name, params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(out), err)
	}
	return nil
}
