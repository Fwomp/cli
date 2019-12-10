/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package http

import (
	"encoding/json"
	"strconv"

	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/version"
)

// Error API errors
type Error struct {
	Err     error
	Message string
	Code    int
}

// Unwrap get the original error
func (e *Error) Unwrap() error { return e.Err }

// IsNil whether the error is nil
func (e *Error) IsNil() bool { return e.Err == nil && e.Message == "" }

func apiKeyHeader(apiKey string) map[string]string {
	return map[string]string{"api-key": apiKey}
}

// GenerateAuthCode generate an auth code
func GenerateAuthCode(host string, verifyTLS bool, hostname string, os string, arch string) (map[string]interface{}, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "hostname", Value: hostname})
	params = append(params, queryParam{Key: "version", Value: version.ProgramVersion})
	params = append(params, queryParam{Key: "os", Value: os})
	params = append(params, queryParam{Key: "arch", Value: arch})

	statusCode, response, err := GetRequest(host, verifyTLS, nil, "/auth/v1/cli/generate", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch auth code", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	return result, Error{}
}

// GetAuthToken get an auth token
func GetAuthToken(host string, verifyTLS bool, code string) (map[string]interface{}, Error) {
	reqBody := map[string]interface{}{}
	reqBody["code"] = code
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, Error{Err: err, Message: "Invalid auth code"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, nil, "/auth/v1/cli/authorize", []queryParam{}, body)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch auth token", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch auth token", Code: statusCode}
	}

	return result, Error{}
}

// RollAuthToken roll an auth token
func RollAuthToken(host string, verifyTLS bool, token string) (map[string]interface{}, Error) {
	reqBody := map[string]interface{}{}
	reqBody["token"] = token
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, Error{Err: err, Message: "Invalid auth token"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, nil, "/auth/v1/cli/roll", []queryParam{}, body)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to roll auth token", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	return result, Error{}
}

// RevokeAuthToken revoke an auth token
func RevokeAuthToken(host string, verifyTLS bool, token string) (map[string]interface{}, Error) {
	reqBody := map[string]interface{}{}
	reqBody["token"] = token
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, Error{Err: err, Message: "Invalid auth token"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, nil, "/auth/v1/cli/revoke", []queryParam{}, body)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to revoke auth token", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	return result, Error{}
}

// DownloadSecrets for specified project and config
func DownloadSecrets(host string, verifyTLS bool, apiKey string, project string, config string, metadata bool) ([]byte, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "project", Value: project})
	params = append(params, queryParam{Key: "metadata", Value: strconv.FormatBool(metadata)})

	headers := apiKeyHeader(apiKey)
	headers["Accept"] = "text/plain"
	statusCode, response, err := GetRequest(host, verifyTLS, headers, "/enclave/v1/configs/"+config+"/secrets", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to download secrets", Code: statusCode}
	}

	return response, Error{}
}

// GetSecrets for specified project and config
func GetSecrets(host string, verifyTLS bool, apiKey string, project string, config string) ([]byte, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "project", Value: project})

	headers := apiKeyHeader(apiKey)
	headers["Accept"] = "application/json"
	statusCode, response, err := GetRequest(host, verifyTLS, headers, "/enclave/v1/configs/"+config+"/secrets", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch secrets", Code: statusCode}
	}

	return response, Error{}
}

// SetSecrets for specified project and config
func SetSecrets(host string, verifyTLS bool, apiKey string, project string, config string, secrets map[string]interface{}) (map[string]models.ComputedSecret, Error) {
	reqBody := map[string]interface{}{}
	reqBody["secrets"] = secrets
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, Error{Err: err, Message: "Invalid secrets"}
	}

	var params []queryParam
	params = append(params, queryParam{Key: "project", Value: project})

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/configs/"+config+"/secrets", params, body)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to set secrets", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	computed := map[string]models.ComputedSecret{}
	for key, secret := range result["secrets"].(map[string]interface{}) {
		val := secret.(map[string]interface{})
		computed[key] = models.ComputedSecret{Name: key, RawValue: val["raw"].(string), ComputedValue: val["computed"].(string)}
	}

	return computed, Error{}
}

// GetWorkplaceSettings get specified workplace settings
func GetWorkplaceSettings(host string, verifyTLS bool, apiKey string) (models.WorkplaceSettings, Error) {
	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/workplace/v1", []queryParam{})
	if err != nil {
		return models.WorkplaceSettings{}, Error{Err: err, Message: "Unable to fetch workplace settings", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.WorkplaceSettings{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	settings := models.ParseWorkplaceSettings(result["workplace"].(map[string]interface{}))
	return settings, Error{}
}

// SetWorkplaceSettings set workplace settings
func SetWorkplaceSettings(host string, verifyTLS bool, apiKey string, values models.WorkplaceSettings) (models.WorkplaceSettings, Error) {
	body, err := json.Marshal(values)
	if err != nil {
		return models.WorkplaceSettings{}, Error{Err: err, Message: "Invalid workplace settings"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/workplace/v1", []queryParam{}, body)
	if err != nil {
		return models.WorkplaceSettings{}, Error{Err: err, Message: "Unable to update workplace settings", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.WorkplaceSettings{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	settings := models.ParseWorkplaceSettings(result["workplace"].(map[string]interface{}))
	return settings, Error{}
}

// GetProjects get projects
func GetProjects(host string, verifyTLS bool, apiKey string) ([]models.ProjectInfo, Error) {
	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/projects", []queryParam{})
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch projects", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	var info []models.ProjectInfo
	for _, project := range result["projects"].([]interface{}) {
		projectInfo := models.ParseProjectInfo(project.(map[string]interface{}))
		info = append(info, projectInfo)
	}
	return info, Error{}
}

// GetProject get specified project
func GetProject(host string, verifyTLS bool, apiKey string, project string) (models.ProjectInfo, Error) {
	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/projects/"+project, []queryParam{})
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to fetch project", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	projectInfo := models.ParseProjectInfo(result["project"].(map[string]interface{}))
	return projectInfo, Error{}
}

// CreateProject create a project
func CreateProject(host string, verifyTLS bool, apiKey string, name string, description string) (models.ProjectInfo, Error) {
	postBody := map[string]string{"name": name, "description": description}
	body, err := json.Marshal(postBody)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Invalid project info"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/projects", []queryParam{}, body)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to create project", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	projectInfo := models.ParseProjectInfo(result["project"].(map[string]interface{}))
	return projectInfo, Error{}
}

// UpdateProject update a project
func UpdateProject(host string, verifyTLS bool, apiKey string, project string, name string, description string) (models.ProjectInfo, Error) {
	postBody := map[string]string{"name": name, "description": description}
	body, err := json.Marshal(postBody)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Invalid project info"}
	}

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/projects/"+project, []queryParam{}, body)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to update project", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ProjectInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	projectInfo := models.ParseProjectInfo(result["project"].(map[string]interface{}))
	return projectInfo, Error{}
}

// DeleteProject create a project
func DeleteProject(host string, verifyTLS bool, apiKey string, project string) Error {
	statusCode, response, err := DeleteRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/projects/"+project, []queryParam{})
	if err != nil {
		return Error{Err: err, Message: "Unable to delete project", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	return Error{}
}

// GetEnvironments get environments
func GetEnvironments(host string, verifyTLS bool, apiKey string, project string) ([]models.EnvironmentInfo, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "project", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/environments", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch environments", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	var info []models.EnvironmentInfo
	for _, environment := range result["environments"].([]interface{}) {
		environmentInfo := models.ParseEnvironmentInfo(environment.(map[string]interface{}))
		info = append(info, environmentInfo)
	}
	return info, Error{}
}

// GetEnvironment get specified environment
func GetEnvironment(host string, verifyTLS bool, apiKey string, project string, environment string) (models.EnvironmentInfo, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "project", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/enclave/v1/environments/"+environment, params)
	if err != nil {
		return models.EnvironmentInfo{}, Error{Err: err, Message: "Unable to fetch environment", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.EnvironmentInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	info := models.ParseEnvironmentInfo(result["environment"].(map[string]interface{}))
	return info, Error{}
}

// GetConfigs get configs
func GetConfigs(host string, verifyTLS bool, apiKey string, project string) ([]models.ConfigInfo, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch configs", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	var info []models.ConfigInfo
	for _, config := range result["environments"].([]interface{}) {
		configInfo := models.ParseConfigInfo(config.(map[string]interface{}))
		info = append(info, configInfo)
	}
	return info, Error{}
}

// GetConfig get a config
func GetConfig(host string, verifyTLS bool, apiKey string, project string, config string) (models.ConfigInfo, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config, params)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to fetch configs", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	info := models.ParseConfigInfo(result["environment"].(map[string]interface{}))
	return info, Error{}
}

// CreateConfig create a config
func CreateConfig(host string, verifyTLS bool, apiKey string, project string, name string, environment string, defaults bool) (models.ConfigInfo, Error) {
	postBody := map[string]interface{}{"name": name, "stage": environment, "defaults": defaults}
	body, err := json.Marshal(postBody)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Invalid config info"}
	}

	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments", params, body)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to create config", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	info := models.ParseConfigInfo(result["environment"].(map[string]interface{}))
	return info, Error{}
}

// DeleteConfig create a config
func DeleteConfig(host string, verifyTLS bool, apiKey string, project string, config string) Error {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := DeleteRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config, params)
	if err != nil {
		return Error{Err: err, Message: "Unable to delete config", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	return Error{}
}

// UpdateConfig create a config
func UpdateConfig(host string, verifyTLS bool, apiKey string, project string, config string, name string) (models.ConfigInfo, Error) {
	postBody := map[string]interface{}{"name": name}
	body, err := json.Marshal(postBody)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Invalid config info"}
	}

	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config, params, body)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to update config", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ConfigInfo{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	info := models.ParseConfigInfo(result["environment"].(map[string]interface{}))
	return info, Error{}
}

// GetActivityLogs get activity logs
func GetActivityLogs(host string, verifyTLS bool, apiKey string) ([]models.Log, Error) {
	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/logs", []queryParam{})
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch activity logs", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	var logs []models.Log
	for _, log := range result["logs"].([]interface{}) {
		parsedLog := models.ParseLog(log.(map[string]interface{}))
		logs = append(logs, parsedLog)
	}
	return logs, Error{}
}

// GetActivityLog get specified activity log
func GetActivityLog(host string, verifyTLS bool, apiKey string, log string) (models.Log, Error) {
	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/logs/"+log, []queryParam{})
	if err != nil {
		return models.Log{}, Error{Err: err, Message: "Unable to fetch activity log", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.Log{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	parsedLog := models.ParseLog(result["log"].(map[string]interface{}))
	return parsedLog, Error{}
}

// GetConfigLogs get config audit logs
func GetConfigLogs(host string, verifyTLS bool, apiKey string, project string, config string) ([]models.ConfigLog, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config+"/logs", params)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to fetch config logs", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	var logs []models.ConfigLog
	for _, log := range result["logs"].([]interface{}) {
		parsedLog := models.ParseConfigLog(log.(map[string]interface{}))
		logs = append(logs, parsedLog)
	}
	return logs, Error{}
}

// GetConfigLog get config audit log
func GetConfigLog(host string, verifyTLS bool, apiKey string, project string, config string, log string) (models.ConfigLog, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := GetRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config+"/logs/"+log, params)
	if err != nil {
		return models.ConfigLog{}, Error{Err: err, Message: "Unable to fetch config log", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ConfigLog{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	parsedLog := models.ParseConfigLog(result["log"].(map[string]interface{}))
	return parsedLog, Error{}
}

// RollbackConfigLog rollback a config log
func RollbackConfigLog(host string, verifyTLS bool, apiKey string, project string, config string, log string) (models.ConfigLog, Error) {
	var params []queryParam
	params = append(params, queryParam{Key: "pipeline", Value: project})

	statusCode, response, err := PostRequest(host, verifyTLS, apiKeyHeader(apiKey), "/v2/environments/"+config+"/logs/"+log+"/rollback", params, []byte{})
	if err != nil {
		return models.ConfigLog{}, Error{Err: err, Message: "Unable to rollback config log", Code: statusCode}
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return models.ConfigLog{}, Error{Err: err, Message: "Unable to parse API response", Code: statusCode}
	}

	parsedLog := models.ParseConfigLog(result["log"].(map[string]interface{}))
	return parsedLog, Error{}
}
