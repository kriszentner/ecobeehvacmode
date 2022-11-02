# ecobeehvacmode

This is a simple Go program I created for a home automation project.

## Set up
My HVAC system at home is a combination heat pump and furnace. My thermostat is an Ecobee.

## Problem
The Ecobee uses an unknown weather provider for their outside temperatures. Unfortunately for many, that weather source can often 2-3 be degrees C off, and there's no way to customize where the weather comes from.

## Solutions

I created a Go program that will, once you get initial authentication set up switch your heat pump or furnace. My main application is using this with homekit. I have an automation that queries my Eve Weather device, and then runs an ssh command to my Homebridge server on a Raspberry Pi where this Go app exists. 

### Caveats
Since Homebridge runs on a older version of Debian, you may need to change a couple functions that use `io` so this works on Go 1.15 (unless you find a newer version of Go for Debian Bullseye on the armv7l platform).

It's theoretically possible to do the Open Weather Map automation (`w` mode) via Home Assistant, but I wasn't able to get it to switch my Ecobee to Aux mode.

## Initial authentication

You'll need to be an [Ecobee Developer](https://www.ecobee.com/en-us/developers/)
- After this you can create an API key by logging into ecobee.com, clicking on the top right, and going to "Developer"
- If there's nothing here, click "Create New" and create an application with the Authorization Method as "ecobee PIN"

Now do the following, using your API key to get the ecobee pin:

```bash
API_KEY="<YOUR API KEY>"
curl -X GET "https://api.ecobee.com/authorize?response_type=ecobeePin&client_id=${API_KEY}&scope=smartWrite"
```
You'll get a response in the below format:
```json
{
  "ecobeePin": "ABCD-EFGIH",
  "code": "a-B2cd3f4GhijklMnOpQR123",
  "interval": 5,
  "expires_in": 900,
  "scope": "openid,offline_access,smartWrite"
}
```

Once you do this you'll want to go to ecobee.com, login to the web portal and click on the 'My Apps' tab." This will bring you to a page where you can add an application by authorizing your ecobeePin. To do this, type/paste your ecobeePin and click 'Validate'. The next screen will display any permissions the app requires and will ask you to click 'Add Application.'

Then you can fetch the First Token to generate your refresh token. Fill in the code from the above query and your API Key:
```bash
CODE="<CODE FROM PREVIOUS QUERY>"
API_KEY="<YOUR API KEY>"
DATA="grant_type=ecobeePin&code=$CODE&client_id=$API_KEY"
URL="https://api.ecobee.com/token"
curl -L -d $DATA -X POST $URL
```

## Use and Maintenance
### Changing your ecobee mode
```bash
ecobeehvacmode -m [heat|cool|auto|off|auxHeatOnly]
```

### Refreshing your Ecobee Token
You'll want to refresh your Ecobee token at least every 30 days, which you can do running `ecobeehvacmode -r`.

### Use OpenWeatherMap to change your mode
This depends on getting an Open Weather Map API KEY, this requires the following environment variables:
- OWM_WEATHER_LOCATION
  Your weather location. Go to https://openweathermap.org/, do "Search City" and find the name that you wish to use.

- OWM_API_KEY
  Your API key from Open Weather Map
  
- FURNACE_LOCKOUT_TEMP
  This is the temperature in C to stop using your furnace (switch to "heat").

- HEATPUMP_LOCKOUT_TEMP
  This is the temperature in C to stop using your heat pump (switch to auxHeatOnly)

Once the above are set up, you can just run in `w` mode, set up via cron or your scheduler of choice.
```bash
ecobeehvacmode -w
```
