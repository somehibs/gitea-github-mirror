package mirror

import (
	"bytes"
	"fmt"
	"net/http"
)

func GHCreate(user, pass, repo, desc string) error {
	createBody := fmt.Sprintf(`{"name":"%s","description":"%s"}`, repo, desc)
	req, err := http.NewRequest("POST", "https://api.github.com/user/repos?access_token="+pass, bytes.NewBuffer([]byte(createBody)))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("status: ", resp.Status)
	//fmt.Println("header: ", resp.Header)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("body: " + string(body))
	return nil
}
