package botserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)





// Posts the given message with the given delay to avoid the
// glitch where the response is before the message in groupme. #ThanksMS
// This is thread blocking.
func (i Instance) PostMessageSync(request Message, delay int) {
	i.postHelper(request, delay)
}
// Posts the given message with the given delay to avoid the
// glitch where the response is before the message in groupme. #ThanksMS
// This is asynchronous and not thread blocking.
func (i Instance) PostMessageAsync(request Message, delay int) {
	go i.postHelper(request, delay)
}

func (i *Instance) postHelper(request Message, delay int) {
	if delay > 0 {
		time.Sleep(time.Duration(int64(time.Millisecond) * int64(delay)))
	}
	jsonVal, err := json.Marshal(request)

	if err != nil {
		i.Log.WithField("err", err).Error("could not convert message request to json")
	}

	i.Log.WithField("json", string(jsonVal)).Info("posting message data to groupme")
	if !i.outputToBuffer {
		response, err := http.Post("https://api.groupme.com/v3/bots/post",
			"application/json", bytes.NewBuffer(jsonVal))

		if err != nil {
			i.Log.WithField("err", err.Error()).Error("send message request unsuccessful")
		} else {
			body, _ := ioutil.ReadAll(response.Body)
			i.Log.WithField("response", string(body)).Info("send message request successful")
		}
	} else {
		i.restfulDebugBuffer.Write(jsonVal)
		i.Log.Info("Wrote RESTful request to buffer")
	}


}

func RegisterImage(file string) (string, error) {
	//extSplit := strings.Split(file, ".")
	//ext := extSplit[len(extSplit)-1]
	//var err error
	//// check file extension
	//if strings.Compare(ext, "jpg") == 0 ||
	//	strings.Compare(ext, "jpeg") == 0 ||
	//	strings.Compare(ext, "gif") == 0 ||
	//	strings.Compare(ext, "png") == 0 {
	//
	//	cmd := exec.Command("curl", "https://image.groupme.com/pictures", "-X", "POST", "-H", "X-Access-Token: "+env.GmToken, "-H", "Content-Type: image/"+ext, "--data-binary", "@"+file)
	//	err = cmd.Run()
	//	if err != nil {
	//		return "", err
	//	} else {
	//		val, err := cmd.Output()
	//		Logger.WithField("response", val).Info("gm image service request successful")
	//		return string(val), err
	//	}
	//} else {
	//	err = errors.New("invalid file extension")
	//	Logger.WithField("err", err.Error()).Error("could not register image to gm image service")
	//	return "", err
	//}
	return "", nil
	// TODO implement
}
