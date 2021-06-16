# Reproxy ip2location plugin

This reproxy plugin add http headers to the request with geo data based on remote IP, obtained from [ip2location](https://www.ip2location.com/) database file

See [reproxy github](https://github.com/umputun/reproxy#plugins-support) for more details about plugins support

## Docker image
```
ghcr.io/negasus/reproxy-ip2location-plugin:latest
```

## An example

Run the demo

```bash
docker-compose up
```

Send the request

```
curl "http://127.0.0.1:80/backend/foo" -H 'X-Ip: 4.0.0.0'
```

See docker-compose logs

```
...
backend               | ----------[ 1 ]----------
backend               | 2021-06-16 12:44:33.9451572 +0000 UTC m=+72.026740101
backend               | [172.19.0.3:41612] GET /foo
backend               |
backend               | Host: 172.19.0.2:2000
backend               | Content-Length: 0
backend               | User-Agent: curl/7.64.1
backend               | Accept: */*
backend               | X-Forwarded-For: 172.19.0.1
backend               | X-Forwarded-Host: 127.0.0.1
backend               | X-Geo-Country: United States of America     <------
backend               | X-Geo-Country-Code: US                      <------
backend               | X-Ip: 4.0.0.0
backend               | X-Real-Ip: 172.19.0.1
backend               | Accept-Encoding: gzip
...
```

## Configuration

```
Usage:
  reproxy-ip2location-plugin [OPTIONS]

Application Options:
  -l, --listen=   listen on host:port (default: 0.0.0.0:8080) [$LISTEN]
  -r, --reproxy=  reproxy plugins endpoint (default: http://127.0.0.1:8081) [$REPROXY]
  -d, --database= database file path [$DATABASE]
  -f, --fields=   fields string, comma-separated. See allowed values at https://github.com/negasus/reproxy-ip2location-plugin (default: CF,CC,REG,CITY) [$FIELDS]
  -p, --prefix=   http header prefix (default: X-Geo-) [$PREFIX]
  -s, --ipsource= if defined, the remote address will be taken from that http header. For example: X-Real-IP [$IPSOURCE]
  -m, --inmemory  if true, the database file will be loaded to the application memory [$INMEMORY]

Help Options:
  -h, --help      Show this help message
```

### `--listen` (`0.0.0.0:8080`)

Listen host:port

### `--reproxy` (`http://127.0.0.1:8081`)

Reproxy plugin endpoint, defined in the reproxy configuration with `PLUGIN_LISTEN` option

You should enable plugin support in the reproxy configuration with `PLUGIN_ENABLED=true`

### `--database` (`empty value`)

Path to the ip2location database file (binary file format)

### `--fields` (`CF,CC,REG,CITY`)

You can define, which data should be added to the headers

Allowed keys (comma-separated)

```
Key     Description             Header Name Suffix
--------------------------------------------------
CF      country full name       Country
CC      country code            Country-Code
REG     region                  Region
CITY    city                    City
ISP     isp                     Isp
LAT     latitude                Latitude
LON     longitude               Longitude
DOM     domain                  Domain
ZIP     zipcode                 Zipcode
TZ      timezone                Timezone
NS      net speed               Netspeed
IDD     idd code                Iddcode
AREA    area code               Areacode
WEC     weather station code    Weatherstationcode
WEN     weather station name    Weatherstationname
MCC     mcc                     Mcc
MNC     mnc                     Mnc
MB      mobile brand            Mobilebrand
EL      elevation               Elevation
UT      usage type              Usagetype
```

Example: `CF,REG,CIY,UT`

With this option (and default value of `--preifix`) headers will look like
```
X-Geo-Country
X-Geo-Region
X-Geo-City
X-Geo-Usagetype
```

### `--prefix` (`X-Geo-`)

You can redefine string prefix for http header name

An Example:
```
--prefix='X-GeoData-'
```

All headers will look like

```
X-GeoData-Country=United States of America
X-GeoData-Country-Code=US
```

### `--ipsource` (`empty value`)

You can to define http header name, which will be use for obtain the IP address

An example:
```
--ipsource='X-Real-Ip'
```

With this options IP address will be taken from the `X-Real-Ip` http header

### `--inmemory` (`false`)

By default, the Plugin open database file from the disk and seek information with disk IO operations.

If you can use this options, database file will be preloaded to the application memory for avoid disk IO operations

Reason for this option: full ip2location database file has size greater than 1Gb. You can choose not to use this option if you don't have enough memory

## Changelog

### v0.1.0 (2021-06-16)

- initial version