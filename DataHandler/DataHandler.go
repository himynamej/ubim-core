package DataHandler

import (
	"core_api/Configurations"
	"core_api/DataTypes"
	pb "core_api/proto"
	pbfile "core_api/protoFile"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/xeipuuv/gojsonschema"
	con "golang.org/x/net/context"
)

func Protected(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("err:there was an error in token validation")
					}
					return []byte("fafa@443#1543"), nil
				})
				if err != nil {
					json.NewEncoder(w).Encode(DataTypes.Exception{Message: err.Error()})
					return
				}
				if token.Valid {
					context.Set(req, "decoded", token.Claims)
					decoded := context.Get(req, "decoded")
					var login DataTypes.LoginInfo
					mapstructure.Decode(decoded.(jwt.MapClaims), &login)
					//expireDate, err := time.Parse(time.RFC3339, login.Expire)
					if err != nil {
						return
					}
					//expireDate.Sub(time.Now()) > 0
					if true {
						req.Header.Add("username", login.Username)
						next(w, req)
					} else {
						json.NewEncoder(w).Encode(DataTypes.Exception{Message: "err:authorization expired"})
					}
				} else {
					json.NewEncoder(w).Encode(DataTypes.Exception{Message: "err:invalid authorization token"})
				}
			}
		} else {
			json.NewEncoder(w).Encode(DataTypes.Exception{Message: "err:an authorization header is required"})
		}
	})
}
func Request(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		var request DataTypes.Request

		//decode json
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			res := DataTypes.Response{
				Result: err.Error(),
			}
			json.NewEncoder(w).Encode(&res)
			return
		}
		//json schema valid
		fmt.Println("request: ", request)
		err = ValidJson(request.Src, request.Group, request.Key, request.Payload)
		if err != nil {
			res := DataTypes.Response{
				Message: err.Error(),
				Result:  "error",
			}
			json.NewEncoder(w).Encode(&res)
			return
		}
		//grpc send request

		res1B, err := json.Marshal(request.Payload)
		if err != nil {
			response := DataTypes.Response{
				Payload: nil,
				Result:  "error",
				Message: err.Error(),
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
		}
		req := &pb.NoAuth{Group: request.Group, Key: request.Key, Payload: res1B}
		// valid service
		if Configurations.Client[request.Src] == nil {
			//err = Configurations.GrpcClientReConnect(request.Src)
			//if err != nil {
			response := DataTypes.Response{
				Payload: nil,
				Result:  "error",
				Message: "service " + request.Src + " is offline",
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
			//	}
		}

		str, err := Configurations.Client[request.Src].NormRequest(con.Background(), req)
		if err != nil {

			response := DataTypes.Response{
				Payload: nil,
				Result:  "error",
				Message: err.Error(),
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
		}
		fmt.Print(str)
		//decode payload
		/*var Pay interface{}
		err = json.Unmarshal(str.Payload, Pay)
		if err != nil {
			response := DataTypes.Response{
				Payload: nil,
				Result:  "error",
				Message: err.Error(),
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
		}*/
		//fmt.Println(Pay)

		var result interface{}

		err = json.Unmarshal(str.Payload, &result)
		if err != nil {
			response := DataTypes.Response{
				Payload: nil,
				Result:  "error",
				Message: err.Error(),
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
		}
		//set payload and message
		response := DataTypes.Response{
			Payload: result,
			Result:  str.Result,
			Message: str.Message,
		}
		//return response
		json.NewEncoder(w).Encode(&response)
		return
	}
	json.NewEncoder(w).Encode("error")
	return

}

func GetFile(w http.ResponseWriter, req *http.Request) {
	fmt.Println("error in file")
	if req.Method == "GET" {
		/*	var request DataTypes.Request

			//decode json
			err := json.NewDecoder(req.Body).Decode(&request)
			if err != nil {
				res := DataTypes.Response{
					Result: err.Error(),
				}
				json.NewEncoder(w).Encode(&res)
				return
			}
			//json schema valid
			fmt.Println("request: ", request)
			err = ValidJson(request.Src, request.Group, request.Key, request.Payload)
			if err != nil {
				res := DataTypes.Response{
					Message: err.Error(),
					Result:  "error",
				}
				json.NewEncoder(w).Encode(&res)
				return
			}
			//grpc send request

			res1B, err := json.Marshal(request.Payload)
			if err != nil {
				response := DataTypes.Response{
					Payload: nil,
					Result:  "error",
					Message: err.Error(),
				}
				//return response
				json.NewEncoder(w).Encode(&response)
				return
			}
			req := &pb.NoAuth{Group: request.Group, Key: request.Key, Payload: res1B}
			// valid service
			if Configurations.Client[request.Src] == nil {
				//err = Configurations.GrpcClientReConnect(request.Src)
				//if err != nil {
					response := DataTypes.Response{
						Payload: nil,
						Result:  "error",
						Message: "service " + request.Src + " is offline",
					}
					//return response
					json.NewEncoder(w).Encode(&response)
					return
			//	}
			}

			str, err := Configurations.Client[request.Src].NormRequest(con.Background(), req)
			if err != nil {

				response := DataTypes.Response{
					Payload: nil,
					Result:  "error",
					Message: err.Error(),
				}
				//return response
				json.NewEncoder(w).Encode(&response)
				return
			}
			fmt.Print(str)
			//decode payload
			/*var Pay interface{}
			err = json.Unmarshal(str.Payload, Pay)
			if err != nil {
				response := DataTypes.Response{
					Payload: nil,
					Result:  "error",
					Message: err.Error(),
				}
				//return response
				json.NewEncoder(w).Encode(&response)
				return
			}*/
		//fmt.Println(Pay)
		/*
			var result interface{}

			err = json.Unmarshal(str.Payload, &result)
			if err != nil {
				response := DataTypes.Response{
					Payload: nil,
					Result:  "error",
					Message: err.Error(),
				}
				//return response
				json.NewEncoder(w).Encode(&response)
				return
			}
			//set payload and message
			response := DataTypes.Response{
				Payload: result,
				Result:  str.Result,
				Message: str.Message,
			}
			//return response
			json.NewEncoder(w).Encode(&response)
			return
		*/
		fmt.Println("file")
		vars := mux.Vars(req)
		fileID := vars["fileID"]
		token := vars["token"]
		height := vars["height"]
		width := vars["width"]

		fmt.Println(fileID, token, height, width)

		c := Configurations.ClientFile

		var file []byte
		FileID := &pbfile.FileRequest{FileId: fileID, Token: token} // initialize a pb.Rectangle
		stream, err := c.Download(con.Background(), FileID)
		if err != nil {
			fmt.Print(err)
		}
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("%v.ListFeatures(_) = _, %v", c, err)
				break
			}
			file = append(file, chunk.Content...)
			log.Println(len(chunk.Content))
		}

		_, err = w.Write([]byte(file))
		
		return

		//	err = ioutil.WriteFile("/home/mohammadjavad/1580023823562294516.pdf", file, 0640)
		//	if err != nil {
		//		fmt.Println(err)
		//	}

	}
	json.NewEncoder(w).Encode("error")
	return

}

func ValidJson(src string, group string, key string, payload interface{}) error {
	if Configurations.FileValid[src+group+key+".json"] == nil {
		return errors.New("error in request schema ")
	}
	schemaLoader := Configurations.FileValid[src+group+key+".json"]
	documentLoader := gojsonschema.NewGoLoader(payload)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return errors.New("error in gojsonschema :" + err.Error())

	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")

		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return errors.New("The document is not valid")
	}

	return nil
}
