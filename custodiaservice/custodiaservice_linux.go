package custodiaservice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/docker/docker-credential-helpers/credentials"
	"io"
	"net"
	"net/http"
	"strings"
)

/*
 * Custodia JSON data (simple JSON only)
 */
type CustodiaJSON struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func CredsToJSON(creds *credentials.Credentials) ([]byte, error) {
	js := CustodiaJSON{Type: "simple"}
	value := []string{
		base64.StdEncoding.EncodeToString([]byte(creds.Username)),
		base64.StdEncoding.EncodeToString([]byte(creds.Secret)),
	}
	js.Value = strings.Join(value, ".")
	return json.Marshal(js)
}

func (js *CustodiaJSON) GetValue() (user string, secret string, err error) {
	value := strings.SplitN(js.Value, ".", 2)
	if len(value) != 2 {
		err = errors.New("Invalid length")
	}
	userb, err := base64.StdEncoding.DecodeString(value[0])
	if err != nil {
		return
	}
	secretb, err := base64.StdEncoding.DecodeString(value[1])
	if err != nil {
		return
	}
	return string(userb), string(secretb), nil
}

/*
 * Custodia directory listing
 */

type CustodiaList []string

/*
 * Unix Domain Socket dialer
 */

type UDSDialer struct {
	SocketPath string
}

func (uds *UDSDialer) DialUnixSocket(proto, addr string) (conn net.Conn, err error) {
	return net.Dial("unix", uds.SocketPath)
}

func UnixClient(path string) *http.Client {
	uds := &UDSDialer{SocketPath: path}
	transport := &http.Transport{
		Dial: uds.DialUnixSocket,
	}
	client := &http.Client{Transport: transport}
	return client
}

/*
 * Custodia Service
 */

type CustodiaService struct {
	Client      *http.Client
	BaseURL     string
	ContentType string
}

func NewCustodiaService(socketpath string, baseurl string) (cs *CustodiaService, err error) {
	if !strings.HasPrefix(baseurl, "http://localhost/") {
		err = errors.New("Invalid base url: " + baseurl)
		return nil, err
	}
	if !strings.HasSuffix(baseurl, "/") {
		err = errors.New("Invalid base url: " + baseurl)
		return nil, err
	}
	client := UnixClient(socketpath)
	cs = &CustodiaService{
		Client:      client,
		BaseURL:     baseurl,
		ContentType: "application/json",
	}
	return cs, nil
}

func (cs CustodiaService) DoRequest(method, serverURL string, body io.Reader) (resp *http.Response, err error) {
	var url string

	if serverURL != "" {
		url = cs.BaseURL + base64.StdEncoding.EncodeToString([]byte(serverURL))
	} else {
		url = cs.BaseURL
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", cs.ContentType)
	return cs.Client.Do(req)
}

func (cs CustodiaService) MkCollection() (err error) {
	resp, err := cs.DoRequest("POST", "", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	return nil
}

func (cs CustodiaService) AddCredentials(creds *credentials.Credentials) (resp *http.Response, err error) {
	js, err := CredsToJSON(creds)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(js)
	return cs.DoRequest("PUT", creds.ServerURL, body)
}

func (cs CustodiaService) Add(creds *credentials.Credentials) (err error) {
	var resp *http.Response
	resp, err = cs.AddCredentials(creds)
	if err != nil {
		return err
	}
	switch {
	case resp.StatusCode < 300:
		// ok
		return nil
	case resp.StatusCode == 404:
		/* container does not exist */
		err = cs.MkCollection()
		if err != nil {
			return err
		}
		// retry Add
		resp, err = cs.AddCredentials(creds)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return errors.New(resp.Status)
		} else {
			return nil
		}
	default:
		return errors.New(resp.Status)
	}
}

func (cs CustodiaService) Delete(serverURL string) (err error) {
	resp, err := cs.DoRequest("DELETE", serverURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	return nil
}

func (cs CustodiaService) Get(serverURL string) (user string, secret string, err error) {
	resp, err := cs.DoRequest("GET", serverURL, nil)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	case 200:
		js := CustodiaJSON{}
		err = json.NewDecoder(resp.Body).Decode(&js)
		if err != nil {
			return "", "", nil
		}
		return js.GetValue()
	case 404:
		return "", "", credentials.NewErrCredentialsNotFound()
	default:
		return "", "", errors.New(resp.Status)
	}
}

func (cs CustodiaService) List() (result map[string]string, err error) {
	var names []string
	// we don't have metadata, use empty user name
	var account string = ""

	resp, err := cs.DoRequest("GET", "", nil)
	if err != nil {
		return
	}

	switch resp.StatusCode {
	case 200:
		names = CustodiaList{}
		err = json.NewDecoder(resp.Body).Decode(&names)
		if err != nil {
			return
		}
		break
	case 404:
		// directory not found, assume no secrets
		break
	default:
		err = errors.New(resp.Status)
		return
	}

	result = make(map[string]string)
	for i := 0; i < len(names); i++ {
		var temp []byte
		temp, err = base64.StdEncoding.DecodeString(names[i])
		if err == nil {
			result[string(temp)] = account
		} else {
			// not base64 encoded
			result[string(temp)] = account
		}
	}

	return result, nil
}
