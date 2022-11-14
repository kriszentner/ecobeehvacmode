# ecobeehvacmode

This is a simple Go program I created for a home automation project. It helps programatically switch heat (heat pump) and aux (furnace or electric resistance heating coils ) modes.

## Set up
My HVAC system at home is a combination heat pump and furnace. My thermostat is an Ecobee.

## Problem
The Ecobee uses an unknown weather provider for their outside temperatures. Unfortunately for many, that weather source can often be 2°C-3°C off, and there's no way to customize where the weather comes from (whether it's another weather source or a personal weather station).

## Solutions

I created a Go program that will, once you get initial authentication set up switch your heat pump or furnace. My main application is using this with home assistant. I have an automation that queries my Eve Weather device, and then makes an API call to the ecobeehvac mode API service running on my Homebridge server where this Go app exists.

This app can also monitor Open Weather Map and switch your Ecobee to/from heat/aux when it detects your temperature threshold. You'll need to run this with something like cron. I'd recommend every 15 minutes considering API query limits, and how fast temperature is prone to change in general.

### Caveats
If you decide to run this on Homebridge for Raspberry Pi like I did, note that it has a on older version of Go (1.15), which means you'll need to change a couple functions that use `io`. You're better off using the latest Go version from https://go.dev/dl/. The armv6l release is backwards compatible with armv7l.

It's theoretically possible to do the Open Weather Map automation (`w` mode) via Home Assistant, but I wasn't able to get it to switch my Ecobee to Aux mode.

I also tried getting this working with Homekit initially, but for some reason, the trigger to run on temperature changes didn't seem to work, so I ended up switching to Home Assistant instead. I've designed this app with many methods of solving this problem, so let your creativity reign.

You can get an Eve Weather to run on Home Assistant with a Thread network (likely with a Homepod as a master). It's a matter of unpairing it from the Homepod, having Homekit Controller automatically detecting it. From here you can have the Homekit integration in Home Assistant present the Eve Weather to the Homepod. Just select "sensors" as the only item to present.

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
This depends on getting an [Open Weather Map API KEY](https://openweathermap.org/appid). Running it in this mode requires the following environment variables:
- OWM_WEATHER_LOCATION
  Your weather location. Go to https://openweathermap.org/, do "Search City" and find the name that you wish to use.

- OWM_API_KEY
  Your API key from Open Weather Map
  
- FURNACE_LOCKOUT_TEMP
  This is the temperature in C to stop using your backup heat (furnace or electric resistance heating coils) this will switch to "heat".

- HEATPUMP_LOCKOUT_TEMP
  This is the temperature in C to stop using your heat pump (switch to auxHeatOnly)

As an example, I have my FURNACE_LOCKOUT_TEMP set to 4.5 and my HEATPUMP_LOCKOUT_TEMP set to 1.5.

Once the above are set up, you can just run in `w` mode, set up via cron or your scheduler of choice.
```bash
ecobeehvacmode -w
```

### Using this as a web API
You can also put this into API mode
```bash
ecobeehvacmode -d -p 8081
```
I've also included an install.sh file which will install this as a systemd service that can run daemonized. Once you have this running on a system, you can change hvac modes by querying the service. Some examples:
```bash
curl http://<yourserver>/?hvacmode=auxHeatOnly"
```
or
```bash
curl http://<yourserver>/?hvacmode=heat"
```
