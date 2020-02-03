package Configurations

import (
	pb "core_api/proto"
	pbFile "core_api/protoFile"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/grpc"
)

type Configurations struct {
	Settings
	Operation
}

type (
	Settings struct {
		ServerPort         string   `json:"server_port"`
		AllowedOrigins     []string `json:"allowed_origins"`
		AllowedHeaders     []string `json:"allowed_headers"`
		AllowedMethods     []string `json:"allowed_methods"`
		AllowCredentials   bool     `json:"allow_credentials"`
		Debug              bool     `json:"debug"`
		OptionsPassthrough bool     `json:"options_passthrough"`
		RunTime            string   `json:"run_time"`
	}
)

type Operation map[string]gojsonschema.JSONLoader

var FileValid Operation

// client of service
var Client map[string]pb.ServiceClient

//client file
var ClientFile pbFile.ServiceClient

// src and ip service
var ClientIP = map[string]string{
	"user":  "0.0.0.0:50051",
	"third": "172.18.0.1:50051",
	"file":  "172.18.0.1:50052",
}

type Routes struct {
	RoutePath     string
	RouteFunction func(w http.ResponseWriter, req *http.Request)
	RouteMethods  string
}

var Configs Configurations

func (c *Configurations) Init(cf string) error {
	// reading configurations settings from file
	err := c.Settings.Init(cf)
	if err != nil {
		return err
	}
	//validation
	ValidFile()
	// connect to the client
	err = GrpcClient()
	if err != nil {
		return err
	}
	//return
	return nil
}
func GrpcClient() error {
	client := make(map[string]pb.ServiceClient)
	// connect to the grpc client settings from file
	for i, c := range ClientIP {
		if i == "file" {
			fmt.Println("i: ", i, " c: ", c)
			conn, err := grpc.Dial(c, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
				return err
			}

			ClientFile = pbFile.NewServiceClient(conn)

			continue
		}
		fmt.Println("i: ", i, " c: ", c)
		conn, err := grpc.Dial(c, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
			return err
		}

		client[i] = pb.NewServiceClient(conn)
	}
	Client = client
	return nil
}
func GrpcClientReConnect(src string) error {
	client := make(map[string]pb.ServiceClient)
	// connect to the grpc client settings from file
	conn, err := grpc.Dial(ClientIP[src], grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect: ", err)
		return err
	}

	client[src] = pb.NewServiceClient(conn)

	return nil
}

func (s *Settings) Init(cf string) error {
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(string(data)), &s)
	if err != nil {
		return err
	}

	return nil
}

func ValidFile() {

	var files []string

	root := Configs.RunTime
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".json") {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	m := make(map[string]gojsonschema.JSONLoader)
	for _, file := range files {

		s := strings.Split(file, "/")
		fmt.Println(s[len(s)-3] + s[len(s)-2] + s[len(s)-1])
		m[s[len(s)-3]+s[len(s)-2]+s[len(s)-1]] = gojsonschema.NewReferenceLoader("file://" + file)

		//fmt.Println(FileValid[s[len(s)-2]][s[len(s)-2]])

	}
	FileValid = m
}
