package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type Event struct {
	Name string `json:"eventName"`
}
type Discord struct {
	Embeds []Embeds `json:"embeds"`
}

type Embeds struct {
	Fields []Fields `json:"fields"`
}
type Fields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type ServiceEndpoint struct {
	Count int `json:"count"`
	Value []struct {
		Data struct {
		} `json:"data"`
		ID        string `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		URL       string `json:"url"`
		CreatedBy struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			Links       struct {
				Avatar struct {
					Href string `json:"href"`
				} `json:"avatar"`
			} `json:"_links"`
			ID         string `json:"id"`
			UniqueName string `json:"uniqueName"`
			ImageURL   string `json:"imageUrl"`
			Descriptor string `json:"descriptor"`
		} `json:"createdBy"`
		Description   string `json:"description"`
		Authorization struct {
			Parameters struct {
				Username        string `json:"username"`
				Password        string `json:"password"`
				AssumeRoleArn   string `json:"assumeRoleArn"`
				RoleSessionName string `json:"roleSessionName"`
				ExternalID      string `json:"externalId"`
			} `json:"parameters"`
			Scheme string `json:"scheme"`
		} `json:"authorization"`
		IsShared                         bool   `json:"isShared"`
		IsReady                          bool   `json:"isReady"`
		Owner                            string `json:"owner"`
		ServiceEndpointProjectReferences []struct {
			ProjectReference struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"projectReference"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"serviceEndpointProjectReferences"`
	} `json:"value"`
}

func ListAccessKeys(iamClient iamiface.IAMAPI, iamUser string) (*iam.ListAccessKeysOutput, error) {
	result, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: aws.String(iamUser),
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteAccessKey(iamClient iamiface.IAMAPI, accessKey string, iamUser string) bool {
	_, err := iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(accessKey),
		UserName:    aws.String(iamUser),
	})

	if err != nil {
		return false
	}

	return true
}

func CreateAccessKeys(iamClient iamiface.IAMAPI, iamUser string) (*iam.CreateAccessKeyOutput, error) {
	result, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(iamUser),
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func MakeHttpRequest(adoUser string, adoToken string, httpClient *http.Client, httpMethod string, url string, payload io.Reader) (*http.Response, error) {
	auth := adoUser + ":" + adoToken
	request, err := http.NewRequest(httpMethod, url, payload)
	request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	response, err := httpClient.Do(request)

	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		fmt.Println("Request success")
		return response, nil
	default:
		fmt.Println("Request failure")
		return nil, err
	}
}

func GetSsmParameter(ssmClient ssmiface.SSMAPI, path string) (string, error) {
	fmt.Println("Getting parameter", path)
	param, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		return "", err
	}

	return *param.Parameter.Value, nil
}

func SendDiscordNotification(httpClient *http.Client, url string, message []Fields) {
	fmt.Println("Sending Discord notification")
	body := Discord{
		Embeds: []Embeds{{
			message,
		}},
	}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(body)
	request, err := http.NewRequest("POST", url, payload)
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Println("Request failure")
	}

	switch response.StatusCode {
	case 204:
		fmt.Println("Request success")
	default:
		fmt.Println("Request failure")
	}
}

func LambdaHandler(event Event) (bool, error) {

	var (
		adoOrg                 = os.Getenv("ADO_ORG")
		adoProject             = os.Getenv("ADO_PROJECT")
		adoServiceEndpointName = os.Getenv("ADO_SERVICE_ENDPOINT_NAME")
		adoUserSsm             = os.Getenv("ADO_USER_SSM")
		adoTokenSsm            = os.Getenv("ADO_TOKEN_SSM")
		awsRegion              = os.Getenv("AWS_REGION")
		iamUser                = os.Getenv("IAM_USER_NAME")
		discordWebhookUrlSsm   = os.Getenv("DISCORD_WEBHOOK_URL_SSM")
	)

	if event.Name == "rotate" {
		fmt.Println("Rotate event triggered")
		fmt.Println("Rotating keys for", iamUser)
		fields := []Fields{}
		fields = append(fields, Fields{Name: "Status of Iam rotation", Value: "Process starting", Inline: false})

		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)},
		)

		ssmClient := ssm.New(sess)
		iamClient := iam.New(sess)
		httpClient := &http.Client{}

		fmt.Println("Getting ssm parameters")
		adoUser, err := GetSsmParameter(ssmClient, adoUserSsm)
		adoToken, err := GetSsmParameter(ssmClient, adoTokenSsm)
		discordWebhookUrl, err := GetSsmParameter(ssmClient, discordWebhookUrlSsm)

		if err != nil {
			return false, err
		}

		listKeys, err := ListAccessKeys(iamClient, iamUser)

		if err != nil {
			fields = append(fields, Fields{Name: "Iam keys", Value: "failed to list keys", Inline: false})
		}

		fmt.Println("Removing any access keys if any")
		for _, key := range listKeys.AccessKeyMetadata {
			fmt.Println("Deleting access key", *key.AccessKeyId)
			deleteKey := DeleteAccessKey(iamClient, *key.AccessKeyId, iamUser)
			if !deleteKey {
				fields = append(fields, Fields{Name: "Iam keys", Value: "Failed to delete access key", Inline: false})
			}
			accessKey := fmt.Sprint("Removed access key: ", *key.AccessKeyId)
			fields = append(fields, Fields{Name: "Iam keys", Value: accessKey, Inline: false})
		}

		fmt.Println("Creating new access key")
		createKey, err := CreateAccessKeys(iamClient, iamUser)

		if err != nil {
			fields = append(fields, Fields{Name: "Iam keys", Value: "Failed to create access key", Inline: false})
		} else {
			fields = append(fields, Fields{Name: "Iam keys", Value: "New access key created", Inline: false})
		}

		fmt.Println("Getting service connections details")
		getServiceEndpointUrl := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/serviceendpoint/endpoints?endpointNames=%s&type=AWS&api-version=6.1-preview.4", adoOrg, adoProject, adoServiceEndpointName)
		getServiceEndpoint, err := MakeHttpRequest(adoUser, adoToken, httpClient, "GET", getServiceEndpointUrl, nil)
		defer getServiceEndpoint.Body.Close()
		var serviceConnectionData ServiceEndpoint
		body, err := ioutil.ReadAll(getServiceEndpoint.Body)
		json.Unmarshal(body, &serviceConnectionData)
		payload := serviceConnectionData.Value[0]

		if err != nil {
			fields = append(fields, Fields{Name: "Ado service endpoint", Value: "Unable to get details", Inline: false})
		}

		fmt.Println("Updating service connections details")
		payload.Authorization.Parameters.Username = *createKey.AccessKey.AccessKeyId
		payload.Authorization.Parameters.Password = *createKey.AccessKey.SecretAccessKey
		putServiceEndpointUrl := fmt.Sprintf("https://dev.azure.com/%s/_apis/serviceendpoint/endpoints/%s?api-version=6.1-preview.4", adoOrg, serviceConnectionData.Value[0].ID)
		payloadjson, err := json.Marshal(payload)
		putServiceEndpoint, err := MakeHttpRequest(adoUser, adoToken, httpClient, "PUT", putServiceEndpointUrl, bytes.NewBuffer(payloadjson))
		defer putServiceEndpoint.Body.Close()
		if err != nil {
			fields = append(fields, Fields{Name: "Ado service endpoint", Value: "Unable to update details", Inline: false})
		} else {
			fields = append(fields, Fields{Name: "Ado service endpoint", Value: "Service endpoint updated", Inline: false})
		}

		SendDiscordNotification(httpClient, discordWebhookUrl, fields)
	}
	return true, nil
}

func main() {
	environment := os.Getenv("GO_ENVIRONMENT")
	if environment != "" {
		lambda.Start(LambdaHandler)
	} else {
		event := Event{"rotate"}
		_, lambdaHandlerResponseErr := LambdaHandler(event)
		if lambdaHandlerResponseErr != nil {
			fmt.Println("Error:", lambdaHandlerResponseErr)
		}
	}
}
