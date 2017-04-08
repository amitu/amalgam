package acko

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/juju/errors"
)

func NotifyR2D2(oid, app, okind, ekind string, edata, odata interface{}) error {
	params := make(map[string]interface{})
	params["okind"] = okind
	params["oid"] = oid
	params["ekind"] = ekind
	params["app"] = app
	params["edata"] = edata
	params["odata"] = odata

	d, err := json.Marshal(params)
	if err != nil {
		return errors.Trace(err)
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:8001/events/",
		bytes.NewBuffer(d))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Trace(err)
		}

		return errors.Errorf("send_to_r2d2_error, response: %s", string(respBody))
	}
	return nil
}
