package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// Shortening the import reference name seems to make it a bit easier
	owm "github.com/briandowns/openweathermap"
)

// Go and REST
// http://networkbit.ch/golang-http-client/

// Go and OpenWeatherMap
// http://briandowns.github.io/openweathermap/
// Up to 60 calls a minute

const tokenUri = "https://api.ecobee.com/token"
const thermostatURI = "https://api.ecobee.com/1/thermostat?format=json"

// These are set up as environment vars
// Of course you can hardcode these as const or use vars, and compile them in with something like:
// go build -X main.apiKey=abcdefghijklmnopqrstuvwxyzABCDEF

// API Key from Ecobee
var apiKey string = os.Getenv("API_KEY")

// Ecobee has a fairly consistent refreshtoken once it's set up. Set this below.
var refreshTokenFile = os.Getenv("REFRESHTOKEN")

// File location for the refreshtoken. We'll keep track of it via file.
var refreshTokenFile = os.Getenv("REFRESHTOKENFILE")

// Open Weather Map API Key
var owmApiKey string = os.Getenv("OWM_API_KEY")

// Open Weather Map location
var weatherLocation = os.Getenv("OWM_WEATHER_LOCATION")

// If running in w mode, when to lockout furnace vs heat pump
var furnaceLockoutTempC = os.Getenv("FURNACE_LOCKOUT_TEMP")

var heatpumpLockoutTempC = os.Getenv("HEATPUMP_LOCKOUT_TEMP")

type PinResponse struct {
	EcobeePin string `json:"ecobeePin"`
	Code      string `json:"code"`
	Interval  string `json:"interval"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type StatusResponse struct {
	ThermostatCount int      `json:"thermostatCount"`
	RevisionList    []string `json:"revisionList"`
	StatusList      []string `json:"statusList"`
	Status          Status
}

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ThermostatResponse struct {
	Page struct {
		Page       int `json:"page"`
		TotalPages int `json:"totalPages"`
		PageSize   int `json:"pageSize"`
		Total      int `json:"total"`
	} `json:"page"`
	ThermostatList []struct {
		Identifier     string `json:"identifier"`
		Name           string `json:"name"`
		ThermostatRev  string `json:"thermostatRev"`
		IsRegistered   bool   `json:"isRegistered"`
		ModelNumber    string `json:"modelNumber"`
		Brand          string `json:"brand"`
		Features       string `json:"features"`
		LastModified   string `json:"lastModified"`
		ThermostatTime string `json:"thermostatTime"`
		UtcTime        string `json:"utcTime"`
		Settings       struct {
			HvacMode                            string `json:"hvacMode"`
			LastServiceDate                     string `json:"lastServiceDate"`
			ServiceRemindMe                     bool   `json:"serviceRemindMe"`
			MonthsBetweenService                int    `json:"monthsBetweenService"`
			RemindMeDate                        string `json:"remindMeDate"`
			Vent                                string `json:"vent"`
			VentilatorMinOnTime                 int    `json:"ventilatorMinOnTime"`
			ServiceRemindTechnician             bool   `json:"serviceRemindTechnician"`
			EiLocation                          string `json:"eiLocation"`
			ColdTempAlert                       int    `json:"coldTempAlert"`
			ColdTempAlertEnabled                bool   `json:"coldTempAlertEnabled"`
			HotTempAlert                        int    `json:"hotTempAlert"`
			HotTempAlertEnabled                 bool   `json:"hotTempAlertEnabled"`
			CoolStages                          int    `json:"coolStages"`
			HeatStages                          int    `json:"heatStages"`
			MaxSetBack                          int    `json:"maxSetBack"`
			MaxSetForward                       int    `json:"maxSetForward"`
			QuickSaveSetBack                    int    `json:"quickSaveSetBack"`
			QuickSaveSetForward                 int    `json:"quickSaveSetForward"`
			HasHeatPump                         bool   `json:"hasHeatPump"`
			HasForcedAir                        bool   `json:"hasForcedAir"`
			HasBoiler                           bool   `json:"hasBoiler"`
			HasHumidifier                       bool   `json:"hasHumidifier"`
			HasErv                              bool   `json:"hasErv"`
			HasHrv                              bool   `json:"hasHrv"`
			CondensationAvoid                   bool   `json:"condensationAvoid"`
			UseCelsius                          bool   `json:"useCelsius"`
			UseTimeFormat12                     bool   `json:"useTimeFormat12"`
			Locale                              string `json:"locale"`
			Humidity                            string `json:"humidity"`
			HumidifierMode                      string `json:"humidifierMode"`
			BacklightOnIntensity                int    `json:"backlightOnIntensity"`
			BacklightSleepIntensity             int    `json:"backlightSleepIntensity"`
			BacklightOffTime                    int    `json:"backlightOffTime"`
			SoundTickVolume                     int    `json:"soundTickVolume"`
			SoundAlertVolume                    int    `json:"soundAlertVolume"`
			CompressorProtectionMinTime         int    `json:"compressorProtectionMinTime"`
			CompressorProtectionMinTemp         int    `json:"compressorProtectionMinTemp"`
			Stage1HeatingDifferentialTemp       int    `json:"stage1HeatingDifferentialTemp"`
			Stage1CoolingDifferentialTemp       int    `json:"stage1CoolingDifferentialTemp"`
			Stage1HeatingDissipationTime        int    `json:"stage1HeatingDissipationTime"`
			Stage1CoolingDissipationTime        int    `json:"stage1CoolingDissipationTime"`
			HeatPumpReversalOnCool              bool   `json:"heatPumpReversalOnCool"`
			FanControlRequired                  bool   `json:"fanControlRequired"`
			FanMinOnTime                        int    `json:"fanMinOnTime"`
			HeatCoolMinDelta                    int    `json:"heatCoolMinDelta"`
			TempCorrection                      int    `json:"tempCorrection"`
			HoldAction                          string `json:"holdAction"`
			HeatPumpGroundWater                 bool   `json:"heatPumpGroundWater"`
			HasElectric                         bool   `json:"hasElectric"`
			HasDehumidifier                     bool   `json:"hasDehumidifier"`
			DehumidifierMode                    string `json:"dehumidifierMode"`
			DehumidifierLevel                   int    `json:"dehumidifierLevel"`
			DehumidifyWithAC                    bool   `json:"dehumidifyWithAC"`
			DehumidifyOvercoolOffset            int    `json:"dehumidifyOvercoolOffset"`
			AutoHeatCoolFeatureEnabled          bool   `json:"autoHeatCoolFeatureEnabled"`
			WifiOfflineAlert                    bool   `json:"wifiOfflineAlert"`
			HeatMinTemp                         int    `json:"heatMinTemp"`
			HeatMaxTemp                         int    `json:"heatMaxTemp"`
			CoolMinTemp                         int    `json:"coolMinTemp"`
			CoolMaxTemp                         int    `json:"coolMaxTemp"`
			HeatRangeHigh                       int    `json:"heatRangeHigh"`
			HeatRangeLow                        int    `json:"heatRangeLow"`
			CoolRangeHigh                       int    `json:"coolRangeHigh"`
			CoolRangeLow                        int    `json:"coolRangeLow"`
			UserAccessCode                      string `json:"userAccessCode"`
			UserAccessSetting                   int    `json:"userAccessSetting"`
			AuxRuntimeAlert                     int    `json:"auxRuntimeAlert"`
			AuxOutdoorTempAlert                 int    `json:"auxOutdoorTempAlert"`
			AuxMaxOutdoorTemp                   int    `json:"auxMaxOutdoorTemp"`
			AuxRuntimeAlertNotify               bool   `json:"auxRuntimeAlertNotify"`
			AuxOutdoorTempAlertNotify           bool   `json:"auxOutdoorTempAlertNotify"`
			AuxRuntimeAlertNotifyTechnician     bool   `json:"auxRuntimeAlertNotifyTechnician"`
			AuxOutdoorTempAlertNotifyTechnician bool   `json:"auxOutdoorTempAlertNotifyTechnician"`
			DisablePreHeating                   bool   `json:"disablePreHeating"`
			DisablePreCooling                   bool   `json:"disablePreCooling"`
			InstallerCodeRequired               bool   `json:"installerCodeRequired"`
			DrAccept                            string `json:"drAccept"`
			IsRentalProperty                    bool   `json:"isRentalProperty"`
			UseZoneController                   bool   `json:"useZoneController"`
			RandomStartDelayCool                int    `json:"randomStartDelayCool"`
			RandomStartDelayHeat                int    `json:"randomStartDelayHeat"`
			HumidityHighAlert                   int    `json:"humidityHighAlert"`
			HumidityLowAlert                    int    `json:"humidityLowAlert"`
			DisableHeatPumpAlerts               bool   `json:"disableHeatPumpAlerts"`
			DisableAlertsOnIdt                  bool   `json:"disableAlertsOnIdt"`
			HumidityAlertNotify                 bool   `json:"humidityAlertNotify"`
			HumidityAlertNotifyTechnician       bool   `json:"humidityAlertNotifyTechnician"`
			TempAlertNotify                     bool   `json:"tempAlertNotify"`
			TempAlertNotifyTechnician           bool   `json:"tempAlertNotifyTechnician"`
			MonthlyElectricityBillLimit         int    `json:"monthlyElectricityBillLimit"`
			EnableElectricityBillAlert          bool   `json:"enableElectricityBillAlert"`
			EnableProjectedElectricityBillAlert bool   `json:"enableProjectedElectricityBillAlert"`
			ElectricityBillingDayOfMonth        int    `json:"electricityBillingDayOfMonth"`
			ElectricityBillCycleMonths          int    `json:"electricityBillCycleMonths"`
			ElectricityBillStartMonth           int    `json:"electricityBillStartMonth"`
			VentilatorMinOnTimeHome             int    `json:"ventilatorMinOnTimeHome"`
			VentilatorMinOnTimeAway             int    `json:"ventilatorMinOnTimeAway"`
			BacklightOffDuringSleep             bool   `json:"backlightOffDuringSleep"`
			AutoAway                            bool   `json:"autoAway"`
			SmartCirculation                    bool   `json:"smartCirculation"`
			FollowMeComfort                     bool   `json:"followMeComfort"`
			VentilatorType                      string `json:"ventilatorType"`
			IsVentilatorTimerOn                 bool   `json:"isVentilatorTimerOn"`
			VentilatorOffDateTime               string `json:"ventilatorOffDateTime"`
			HasUVFilter                         bool   `json:"hasUVFilter"`
			CoolingLockout                      bool   `json:"coolingLockout"`
			VentilatorFreeCooling               bool   `json:"ventilatorFreeCooling"`
			DehumidifyWhenHeating               bool   `json:"dehumidifyWhenHeating"`
			VentilatorDehumidify                bool   `json:"ventilatorDehumidify"`
			GroupRef                            string `json:"groupRef"`
			GroupName                           string `json:"groupName"`
			GroupSetting                        int    `json:"groupSetting"`
			FanSpeed                            string `json:"fanSpeed"`
		} `json:"settings"`
	} `json:"thermostatList"`
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

func main() {
	var hvacMode string
	var refresh bool
	var weather bool
	flag.StringVar(&hvacMode, "m", "", "hvac mode: heat, cool, auto, off, auxHeatOnly")
	flag.BoolVar(&refresh, "r", false, "Refresh token only")
	flag.BoolVar(&weather, "w", false, "Check and change hvacMode based on OpenWeatherMap")
	flag.Parse()

	refreshToken := readRefreshToken()

	if !refresh {
		if !(hvacMode == "heat" || hvacMode == "cool" || hvacMode == "auto" || hvacMode == "off" || hvacMode == "auxHeatOnly") {
			log.Fatal("Invalid HVAC mode. Must be one of: heat, cool, auto, off, auxHeatOnly")
		} else {
			tokenResponse := renewAccessToken(refreshToken)
			setHvacMode(tokenResponse.AccessToken, hvacMode)
		}
	} else if weather {
		changeBasedOnOwm(refreshToken)
	} else {
		renewAccessToken(refreshToken)
	}
	fmt.Println("done")
}

// Access: New tokens are good for 1 hour
// Refresh: Need to refresh every 30 days to keep this alive
// https://www.ecobee.com/home/developer/api/documentation/v1/auth/token-refresh.shtml
func renewAccessToken(refreshToken string) TokenResponse {

	data := url.Values{
		"grant_type": {"refresh_token"},
		"code":       {refreshToken},
		"client_id":  {apiKey},
	}
	resp, err := http.PostForm(tokenUri, data)

	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	writeRefreshToken(tokenResponse.RefreshToken)
	return (tokenResponse)
}

// Send an API call to change the HVAC mode of the Ecobee
func setHvacMode(accessToken string, hvacMode string) {

	setHvacModeData := []byte(`{"selection":{"selectionType":"registered","selectionMatch":""},"thermostat":{"settings":{"hvacMode":"` + hvacMode + `"}}}`)

	req, err := http.NewRequest("POST", thermostatURI, bytes.NewBuffer(setHvacModeData))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	fmt.Printf("%s\n", body)
}

func getStatusId(accessToken string) string {
	detailsUrl := "https://api.ecobee.com/1/thermostatSummary?format=json&body=%7B%22selection%22%3A%7B%22selectionType%22%3A%22registered%22%2C%22selectionMatch%22%3A%22%22%2C%22includeRuntime%22%3Atrue%2C%22includeSensors%22%3Atrue%2C%22includeSettings%22%3Atrue%2C%22includeEquipmentStatus%22%3Atrue%7D%7D"
	req, err := http.NewRequest("GET", detailsUrl, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	fmt.Printf("%s\n", body)
	var statusResponse StatusResponse
	err = json.Unmarshal(body, &statusResponse)
	if err != nil {
		log.Fatal("Error parsing status response. ", err)
	}
	fmt.Println(statusResponse.RevisionList[0])
	ids := strings.Split(statusResponse.RevisionList[0], ":")

	return ids[0]
}

func getHvacMode(accessToken string, statusId string) string {
	detailsUrl := "https://api.ecobee.com/1/thermostat?json=%7B%0A%20%20%22selection%22%3A%20%7B%0A%20%20%20%20%22selectionType%22%3A%20%22thermostats%22%2C%0A%20%20%20%20%22selectionMatch%22%3A%20%22415514179919%22%2C%0A%20%20%20%20%22includeSettings%22%3A%20%22true%22%0A%20%20%7D%0A%7D"

	req, err := http.NewRequest("GET", detailsUrl, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	// fmt.Printf("%s\n", body)

	var thermostatResponse ThermostatResponse
	err = json.Unmarshal(body, &thermostatResponse)
	if err != nil {
		log.Fatal("Error parsing status response. ", err)
	}

	resp.Body.Close()
	return thermostatResponse.ThermostatList[0].Settings.HvacMode
}

func readRefreshToken() string {
	b, err := os.ReadFile(refreshTokenFile)
	// If the file isn't found, exit early and return the known refreshtoken
	// that doesn't seem to change
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("refreshtoken file missing, returning known refreshtoken")
		return refreshTokenConst
	}
	if err != nil {
		log.Fatal("Error reading file. ", err)
	}
	refreshToken := string(b)

	return refreshToken
}

// Write the refresh token to a file
func writeRefreshToken(refreshToken string) {
	f, err := os.Create(refreshTokenFile)
	if err != nil {
		log.Fatal("Error writing file. ", err)
	}
	defer f.Close()

	f.WriteString(refreshToken)
}

func changeBasedOnOwm(refreshToken string) {
	currentTemp := getTemp()
	if currentTemp < heatpumpLockoutTempC {
		tokenResponse := renewAccessToken(refreshToken)
		statusId := getStatusId(tokenResponse.AccessToken)
		hvacMode := getHvacMode(tokenResponse.AccessToken, statusId)
		if hvacMode == "heat" {
			setHvacMode(tokenResponse.AccessToken, "auxHeatOnly")
		}
	} else if currentTemp > furnaceLockoutTempC {
		tokenResponse := renewAccessToken(refreshToken)
		statusId := getStatusId(tokenResponse.AccessToken)
		hvacMode := getHvacMode(tokenResponse.AccessToken, statusId)
		if hvacMode == "auxHeatOnly" {
			setHvacMode(tokenResponse.AccessToken, "heat")
		}
	}
}

func getTemp() float64 {
	w, err := owm.NewCurrent("C", "EN", owmApiKey) // (internal - OpenWeatherMap reference for Celcius with English output
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByName(weatherLocation)
	fmt.Println(w.Main.Temp)
	return w.Main.Temp
}
