package botserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type BasicHookTest struct {
	validity bool
}

func (a *BasicHookTest) Action(callback Callback) bool {
	print("Action run\n")
	a.validity = true
	return false
}

func (a *BasicHookTest) Name() string {
	return "Basic hook"
}

func QuoteHandler(callback Callback) bool {
	return false
}

////////////////////////////////////////////////////

////////////////////////////////////////////////////

func TestRegisterChannel(t *testing.T) {
	srv := NewInstance()

	basicValidHook := BasicHook{DebugName: "Basic hook", Handler: QuoteHandler}
	noHooks := Channel{
		Name:     "No hook channel",
		GroupIDs: []string{"1234", "2345"},
	}

	noIds := Channel{
		Name:     "No groups channel",
		GroupIDs: []string{},
	}
	noIds.AddHook(&basicValidHook)

	valid := Channel{
		Name:     "Valid Channel",
		GroupIDs: []string{"1234"},
	}
	valid.AddHook(&basicValidHook)

	if srv.RegisterChannel(&noHooks) {
		t.Errorf("no hook channel accepted: %v", noHooks)
	}
	if srv.RegisterChannel(&noIds) {
		t.Errorf("Expected channel to be denied since no ids given. Channel accepted")
	}
	if !srv.RegisterChannel(&valid) {
		t.Errorf("Expected channel to be accepted. Channel denied.")
	}
}

func TestChannel_RemoveHook(t *testing.T) {

}

func TestListenAndServe(t *testing.T) {
	srv := NewInstance()

	basicValidHook := BasicHookTest{}

	validChannel := Channel{
		Name:     "Valid Channel",
		GroupIDs: []string{"30154628"},
	}
	validChannel.AddHook(&basicValidHook)

	// Register channel
	srv.RegisterChannel(&validChannel)

	ts := httptest.NewServer(http.HandlerFunc(srv.requestHandler))
	defer ts.Close()

	request := httptest.NewRequest("POST", "http://example.com", bytes.NewBuffer([]byte("{\"attachments\":[],\"avatar_url\":\"https://i.groupme.com/750x750.jpeg.033de3766a1c414b99f9aad614937c7d\",\"created_at\":1536094557,\"group_id\":\"30154628\",\"id\":\"153609455785477567\",\"name\":\"Neo Featherman\",\"sender_id\":\"31212732\",\"sender_type\":\"user\",\"source_guid\":\"2CE1796D-EA69-4609-8B1A-3058306EAAA6\",\"system\":false,\"text\":\"Do worker bees get compensated for accidents on the job?\",\"user_id\":\"31212732\"}")))
	writer := httptest.NewRecorder()
	srv.requestHandler(writer, request)
	time.Sleep(100 * time.Millisecond)

	log.Print(writer.Body)

	if !basicValidHook.validity {
		t.Error("hook validity test failed")
	}
}

func TestFireHook(t *testing.T) {
	srv := NewInstance()

	// create basic hook and register
	basicValidHook := BasicHookTest{}
	validChannel := Channel{
		Name:     "Valid Channel",
		GroupIDs: []string{"1234"},
	}
	validChannel.AddHook(&basicValidHook)

	callback := Callback{GroupID: "1234"}

	// Register channel
	srv.RegisterChannel(&validChannel)

	for _, e := range srv.channels {
		e.inputCallback(callback)
	}

	log.Println("Waiting for process to end")

	time.Sleep(10 * time.Millisecond)

	if !basicValidHook.validity {
		t.Errorf("validity failure")
	}
}
