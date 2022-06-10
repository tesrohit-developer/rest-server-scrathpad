package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tesrohit-developer/rest-server-scrathpad/plugin"
	"github.fkinternal.com/tupili-easwar/sideline-em-models/models"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	gplugin "github.com/hashicorp/go-plugin"
	emclientmodels "github.fkinternal.com/Flipkart/entity-manager/modules/entity-manager-client-model/EntityManagerClientModel"
	emmodels "github.fkinternal.com/Flipkart/entity-manager/modules/entity-manager-model/EntityManagerModel"
	serde "github.fkinternal.com/Flipkart/entity-manager/modules/entity-manager-schema-registry-go-client"
	"google.golang.org/protobuf/proto"
)

type SidelineEm struct{}

func execute(method, url string, headers map[string]string, payload io.Reader) (bool, int, emclientmodels.ResponseCode, int32, string) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          2, // TODO
			MaxIdleConnsPerHost:   2, // TODO
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 10 * time.Second, // TODO
	}

	//Never fail always recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered in execute %s %v", url, r)
		}
	}()

	//build request
	request, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Printf("failed in request build %s %s \n", url, err.Error())
		return false, 1, -1, 0, ""
	}

	//set headers
	for key, val := range headers {
		request.Header.Set(key, val)
	}

	// if method != "GET" && !h.conf.CustomURL {
	// 	request.Header.Set("Content-Type", "application/octet-stream")
	// }

	//make request
	response, err := client.Do(request)
	if err != nil {
		log.Printf("failed in http call invoke %s %s \n", url, err.Error())
		return false, 2, -1, 0, ""
	}
	//TODO check if this can be avoided
	responseBytes, _ := ioutil.ReadAll(response.Body)
	responseStatusCode := response.StatusCode
	if responseStatusCode > 300 {
		return false, responseStatusCode, emclientmodels.ResponseCode_UNKNOWN, 0, ""
	}
	var readResponse emclientmodels.ReadEntityResponse
	// readResponse.Entity
	proto.Unmarshal(responseBytes, &readResponse)
	emReadResponseCode := readResponse.ResponseMeta.ResponseCode
	io.Copy(ioutil.Discard, response.Body)
	defer response.Body.Close()
	if emclientmodels.ResponseStatus_STATUS_SUCCESS.Number() == readResponse.ResponseMeta.ResponseStatus.Number() {
		return true, responseStatusCode, emReadResponseCode, readResponse.Version, readResponse.String()
	}
	return false, responseStatusCode, emReadResponseCode, readResponse.Version, readResponse.String()
}

func (SidelineEm) CheckMessageSideline(b []byte) ([]byte, error) {
	fmt.Println("Checking message in EM")
	url := "http://10.24.19.136/entity-manager/v1/entity/read"
	headers := make(map[string]string)
	headers["Content-Type"] = "application/octet-stream"
	headers["X-IDEMPOTENCY-ID"] = time.Now().String()
	headers["X-CLIENT-ID"] = "go-dmux"
	//headers["X-PERF-TTL"] = "LONG_PERF"
	entityIdentifier := emmodels.EntityIdentifier{
		Namespace: "com.dmux",
		Name:      "SidelineMessage",
	}
	tenantIdentifier := emmodels.TenantIdentifier{
		Name: "OMSDMUX",
	}
	readEntityRequest := emclientmodels.ReadEntityRequest{
		EntityIdentifier: &entityIdentifier,
		TenantIdentifier: &tenantIdentifier,
		EntityId:         "OD39848785211959690",
		FieldsToRead:     nil,
	}
	b, e := proto.Marshal(&readEntityRequest)
	if e != nil {
		fmt.Println("error in ser ReadEntityRequest")
		return nil, errors.New("error in ser ReadEntityRequest")
	}
	responseBoolean, responseCode, emResponseCode, _, readResponseString := execute("POST", url, headers, bytes.NewReader(b))
	if !responseBoolean {
		if emclientmodels.ResponseCode_ENTITY_NOT_FOUND.Number() == emResponseCode.Number() {
			fmt.Println("Not sidelined message ")
			return nil, nil
		}
		fmt.Println("error in reading Sideline Table")
		errStr := "error in reading Sideline Table, ResponseCode: " + strconv.Itoa(responseCode) +
			" EmResponseCode: " + emResponseCode.String() +
			" ResponseBoolean: " + strconv.FormatBool(responseBoolean) +
			" ReadResponseString: " + readResponseString
		return nil, errors.New(errStr)
	}
	if responseCode < 300 {
		fmt.Println("Success ")
		return nil, nil
	}
	return nil, errors.New("error in reading Sideline Table")
}

func (SidelineEm) SidelineMessage(msg []byte) error {
	// do nothing
	fmt.Println("Sidelining message in EM")
	var message = models.SidelineMessage{
		GroupId:           "G1",
		Partition:         1,
		Message:           nil,
		Offset:            1,
		ConsumerGroupName: "G1",
		ClusterName:       "OMSDMUX",
		EntityId:          "1",
		Version:           1,
	}
	entityIdentifier := emmodels.EntityIdentifier{
		Namespace: "com.dmux",
		Name:      "SidelineMessage",
	}
	tenantIdentifier := emmodels.TenantIdentifier{
		Name: "OMSDMUX",
	}
	upsertEntityRequest := emclientmodels.UpsertEntityRequest{
		EntityIdentifier: &entityIdentifier,
		TenantIdentifier: &tenantIdentifier,
		Entity:           serde.Serialize(message.ProtoReflect()),
		RequestMeta:      nil,
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/octet-stream"
	headers["X-IDEMPOTENCY-ID"] = message.String()
	headers["X-CLIENT-ID"] = "go-dmux"
	headers["X-PERF-TTL"] = "LONG_PERF"
	url := "http://10.24.19.136/entity-manager/v1/entity/upsert"
	b, e := proto.Marshal(&upsertEntityRequest)
	if e != nil {
		return errors.New("error in ser UpsertEntityRequest")
	}
	execute("POST", url, headers, bytes.NewReader(b))
	return nil
}

type SidelineEmPlugin struct{}

func (SidelineEmPlugin) Server(*gplugin.MuxBroker) (interface{}, error) {
	return &plugin.CheckMessageSidelineRPCServer{Impl: new(SidelineEm)}, nil
}

func (SidelineEmPlugin) Client(b *gplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &plugin.CheckMessageSidelineRPC{Client: c}, nil
}

func main() {
	// We're a plugin! Serve the plugin. We set the handshake config
	// so that the host and our plugin can verify they can talk to each other.
	// Then we set the plugin map to say what plugins we're serving.
	gplugin.Serve(&gplugin.ServeConfig{
		HandshakeConfig: gplugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "BASIC_PLUGIN",
			MagicCookieValue: "hello",
		},
		Plugins: pluginMap,
	})
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]gplugin.Plugin{
	"em": new(SidelineEmPlugin),
}
