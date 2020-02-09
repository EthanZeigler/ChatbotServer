package botserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)
type Instance struct {
	// FIXME room for a race condition here that will need to be fixed in the future
	// Internal map from group ids to the linked channels
	groupIDMap map[string][]Channel
	channels []Channel
	outputToBuffer bool
	restfulDebugBuffer bytes.Buffer

	// Conf stores configuration values from the local config.json file
	Config Configuration

	// Logger is the internal server logger made accessible for more complex requests
	// powered by logrus.
	Log *logrus.Logger
}

func (i *Instance) init() {
	i.groupIDMap = make(map[string][]Channel)
	i.channels = make([]Channel, 5)
	i.Log = logrus.New()
}

func NewInstance() (instance *Instance) {
	instance = new(Instance)
	instance.init()
	return
}

// Configuration stores data from the config.json file
type Configuration struct {
	// GroupMe API access token
	GmToken string `json:"gm_token"`
	// Port to bind the server to
	Port string `json:"port"`
	// Additional information inside the config file for individual bot configurations
	BotConfig map[string]interface{} `json:"bot"`
	// Default State? (wasn't loaded)
	initialized bool
}

// Message represents a message to be sent to a groupme group chat.
// Use MakeMessage to initialize
type Message struct {
	// id of the bot to send the message
	BotID string `json:"bot_id"`
	// Text of the message. This is optional, but then must have a picture.
	Text string `json:"text,omitempty"`
	// Picture of the message. This is optional, but then must have text.
	// The provided URL must be from the groupme message API.
	Picture string `json:"picture_url,omitempty"`
}

// initial details to load before code execution including the logger
func (i *Instance) ConfigureFromFile(confFile string) error {
	i.Log = logrus.New()
	i.Log.Info("loading configuration file")
	file, err := ioutil.ReadFile(confFile)
	if err != nil {
		i.Log.WithFields(logrus.Fields{
			"file": confFile,
			"err": err,
		}).Error("file does not exist")
		return err
	}
	
	if err = json.Unmarshal(file, &i.Config); err != nil {
		i.Log.WithFields(logrus.Fields{
			"config": i.Config,
			"err": err,
		}).Error("invalid config")
		return err
	}

	// log loaded config.json file
	i.Log.WithFields(logrus.Fields{
		"gm_token":   i.Config.GmToken,
		"port":       i.Config.Port,
		"bot_config": i.Config.BotConfig,
	}).Info("loaded configuration")
	i.Config.initialized = true

	return nil
}

// Start opens the tcp port and begins listening for input.
// This is blocking for the runtime of the server
func (i *Instance) Start() error {
	// default initialization failed?
	if !i.Config.initialized && i.ConfigureFromFile("config.json") != nil {
		i.Log.Fatal("default initialization failed")
		return errors.New("default initialization failed")
	}
	http.HandleFunc("/", i.requestHandler)
	err := http.ListenAndServe(":" + i.Config.Port, nil)
	i.Log.WithField("err", err).Fatal("web server failed")
	return err
}


func (i *Instance) requestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// not even worth logging
		return
	}

	var callback Callback
	err = json.Unmarshal(body, &callback)
	if err != nil {
		// log marshal failure
		i.Log.WithFields(logrus.Fields{
			"body":   body,
			"source": r.RemoteAddr,
		}).Warn("marshal failure")
		return
	}

	i.Log.WithField("callback", callback).Info("received callback")

	if channels, exists := i.groupIDMap[callback.GroupID]; exists {
		for _, channel := range channels {
			i.Log.WithFields(logrus.Fields{
				"channel": channel.Name,
			}).Info("alerted channel to callback")
			channel.inputCallback(callback)
		}
	} else {
		i.Log.WithFields(logrus.Fields{
			"group_id": callback.GroupID,
		}).Warn("request from unserved group id")
		return
	}
}

// RegisterChannel registers a channel to serve the GM bot IDs and group IDs it defines
// returns true if channel was registered
func (i *Instance) RegisterChannel(channel *Channel) bool {
	// Inspect channel for potential mistakes

	// check that the group has registered group IDs
	if len(channel.GroupIDs) < 1 {
		i.Log.WithFields(logrus.Fields{
			"channel": channel,
			"ids":     channel.GroupIDs,
		}).Warn("channel has no group ids and will not register")
		return false
	}

	// check that the channel has registered hooks
	if len(channel.hooks) < 1 {
		i.Log.WithFields(logrus.Fields{
			"channel": channel,
			"hooks":   channel.hooks,
		}).Warn("channel has no hooks in channel and will not register")
		return false
	}

	// The channel is validated
	// We can now register it into the system
	channel.log = i.Log

	// add to the list of all channels
	i.channels = append(i.channels, *channel)
	// add into the map storing each channel to IDs it listens to
	// speed up lookup
	for _, groupID := range channel.GroupIDs {
		i.groupIDMap[groupID] = append(i.groupIDMap[groupID], *channel)
	}

	i.Log.WithFields(logrus.Fields{
		"name":       channel.Name,
		"groups":     channel.GroupIDs,
		"hook_count": len(channel.hooks),
	}).Info("channel registered")
	return true
}