# ATC Api

ATC API is a Django app that allow to bridge a REST API to ATCD's thrift API.

# Setup

## Requirements

* [Django 1.7](https://github.com/django/django)
* [Django REST framework 3.X](https://github.com/tomchristie/django-rest-framework)
* [atc_thrift](../atc_thrift)

## Installation

The easiest way to install `django-atc-api` is to install it directly from [pip](https://pypi.python.org/pypi).

### From pip
```bash
pip install django-atc-api
```
### From source

```bash
$ cd path/to/django-atc-api
pip install .
```

## Configuration

1. Edit your Django project's `settings.py` and add `atc_api` and `rest_framework` to your `INSTALLED_APPS`:

```python
    INSTALLED_APPS = (
        ...
        'atc_api',
        'rest_framework',
    )
```

2. Include the `atc_api` URLconf in your Django project urls.py like this:

```python
    url(r'^api/v1/', include('atc_api.urls')),
```

3. Start the development server

```bash
python manage.py runserver 0.0.0.0:8000
```

4. Visit http://127.0.0.1:8000/api/v1/shape/ to set/unset shaping.


Some settings like the `ATCD_HOST` and `ATCD_PORT` can be changes in your Django project'settings.py:

```python
ATC_API = {
    'ATCD_HOST': 'localhost',
    'ATCD_PORT': 9090,
}
```

see [ATC api settings](atc_api/settings.py) for more details.


# API usage

Let's suppose the api is available under `/api/v1`. The core API is limited and allow to:

* Getting the shaping staus of an device by GETing `/api/v1/shape/`
* Shape a device by POSTing to `/api/v1/shape/`
* Unshape a device by sending a DELETE request to `/api/v1/shape/`

## Shaping Status

To find out if a device is shaped, you can GET `/api/v1/shape/[ip/]`

If the device is being shaped, HTTP will return 200 and the current shaping of the device.

If the device is not being shaped, HTTP will return code 404.

Examples:

* Check if I am being shaped (device not being shaped, HTTP code 404):

```sh
$ curl -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
{
  "detail": "This IP (10.0.2.2) is not being shaped"
}
```

* Check if I am being shaped (device being shaped, HTTP code 200):

```sh
$ curl -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
{
  "down": {
    "rate": 400,
    "loss": {
      "percentage": 5.0,
      "correlation": 0.0
    },
    "delay": {
      "delay": 15,
      "jitter": 0,
      "correlation": 0.0
    },
    "corruption": {
      "percentage": 0.0,
      "correlation": 0.0
    },
    "reorder": {
      "percentage": 0.0,
      "correlation": 0.0,
      "gap": 0
    }
  },
  "up": {
    "rate": 200,
    "loss": {
      "percentage": 1.0,
      "correlation": 0.0
    },
    "delay": {
      "delay": 10,
      "jitter": 0,
      "correlation": 0.0
    },
    "corruption": {
      "percentage": 0.0,
      "correlation": 0.0
    },
    "reorder": {
      "percentage": 0.0,
      "correlation": 0.0,
      "gap": 0
    }
  }
}
```

* Check if 1.1.1.1 is being shaped (device not being shaped, HTTP code 404):

```sh
$ curl -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/1.1.1.1/
{
  "detail": "This IP (1.1.1.1) is not being shaped"
}
```

## Shaping a device

Shaping a device is done by posting the shaping setting payload to `/api/v1/shape/[ip/]`


Examples:

* Shape my own device, 200kb up, added latency of 10ms with 1% packet loss and 400kb down with added latency of 15ms and 5% packet loss
This will always retun HTTP code 201 on success. If the device was already being shaped, the new setting is going to be applied and the onld one deleted.

Mind the (Ctrl-D)

```sh
$ curl -X POST -d '@-' -i -H 'Content-Type: application/json' -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
{
    "down": {
        "rate": 400,
        "loss": {
            "percentage": 5.0,
            "correlation": 0.0
        },
        "delay": {
            "delay": 15,
            "jitter": 0,
            "correlation": 0.0
        },
        "corruption": {
            "percentage": 0.0,
            "correlation": 0.0
        },
        "reorder": {
            "percentage": 0.0,
            "correlation": 0.0,
            "gap": 0
        }
    },
    "up": {
        "rate": 200,
        "loss": {
            "percentage": 1.0,
            "correlation": 0.0
        },
        "delay": {
            "delay": 10,
            "jitter": 0,
            "correlation": 0.0
        },
        "corruption": {
            "percentage": 0.0,
            "correlation": 0.0
        },
        "reorder": {
            "percentage": 0.0,
            "correlation": 0.0,
            "gap": 0
        }
    }
}
Ctrl-D
HTTP/1.1 201 CREATED
Server: gunicorn/19.2.1
Date: Fri, 27 Feb 2015 20:02:05 GMT
Connection: close
Transfer-Encoding: chunked
Vary: Accept, Cookie
Content-Type: application/json; indent=2
Allow: GET, POST, DELETE, HEAD, OPTIONS

{
  "down": {
    "rate": 400,
    "loss": {
      "percentage": 5.0,
      "correlation": 0.0
    },
    "delay": {
      "delay": 15,
      "jitter": 0,
      "correlation": 0.0
    },
    "corruption": {
      "percentage": 0.0,
      "correlation": 0.0
    },
    "reorder": {
      "percentage": 0.0,
      "correlation": 0.0,
      "gap": 0
    }
  },
  "up": {
    "rate": 200,
    "loss": {
      "percentage": 1.0,
      "correlation": 0.0
    },
    "delay": {
      "delay": 10,
      "jitter": 0,
      "correlation": 0.0
    },
    "corruption": {
      "percentage": 0.0,
      "correlation": 0.0
    },
    "reorder": {
      "percentage": 0.0,
      "correlation": 0.0,
      "gap": 0
    }
  }
}
```

or... more simply:

```sh
$ curl -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
{
    "down": {
        "rate": 400, 
        "loss": {
            "percentage": 5.0
        }, 
        "delay": {
            "delay": 15
        }, 
        "corruption": {}, 
        "reorder": {}
    }, 
    "up": {
        "rate": 200, 
        "loss": {
            "percentage": 1.0
        }, 
        "delay": {
            "delay": 10
        }, 
        "corruption": {}, 
        "reorder": {}
    }
}
CTRL-D
... same response...
```

Likely, device 1.1.1.1 could be shaped by using URL http://127.0.0.1:8080/api/v1/shape/1.1.1.1/ instead.

## Unshaping a device

Unshaping a device is done by sending a DELETE request to `/api/v1/shape/[ip]/`

Examples:

* Unshape myself (device being shaped, HTTP code 204)

```sh
$ curl -X DELETE -i -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
HTTP/1.1 204 NO CONTENT
Server: gunicorn/19.2.1
Date: Fri, 27 Feb 2015 19:46:58 GMT
Connection: close
Vary: Accept, Cookie
Content-Length: 0
Allow: GET, POST, DELETE, HEAD, OPTIONS

```

* Unshape myself (device not being shaped, HTTP code 400):

```sh
$ curl -X DELETE -i -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/
HTTP/1.1 400 BAD REQUEST
Server: gunicorn/19.2.1
Date: Fri, 27 Feb 2015 19:43:36 GMT
Connection: close
Transfer-Encoding: chunked
Vary: Accept, Cookie
Content-Type: application/json; indent=2
Allow: GET, POST, DELETE, HEAD, OPTIONS

{
  "detail": "{'message': 'No session for IP 10.0.2.2 found', 'result': 12}"
}
```

* Unshape 1.1.1.1 (device not being shaped, HTTP code 400):

```sh
$ curl -X DELETE -i -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/shape/1.1.1.1/
HTTP/1.1 400 BAD REQUEST
Server: gunicorn/19.2.1
Date: Fri, 27 Feb 2015 19:47:57 GMT
Connection: close
Transfer-Encoding: chunked
Vary: Accept, Cookie
Content-Type: application/json; indent=2
Allow: GET, POST, DELETE, HEAD, OPTIONS

{
  "detail": "{'message': 'No session for IP 1.1.1.1 found', 'result': 12}"
}
```

## Authentication and Authorization

ATC employs a token-based authentication system to allow devices to securely shape others.

To use this system, the controlled device must ask for a token from ATC. Once a token is obtained,

the controlling device can post this token to ATC to authorize itself to shape the device.

### Retrieving a Token

Use the `/api/v1/token/` endpoint to retrieve a token.

This endpoint will use the HTTP Header `HTTP_X_REAL_IP` to generate the token.

For security reasons this is the only way to set the client IP. See [Proxy Setup](#proxy-security) below.

```sh
$ curl -i -H 'Accept: application/json; indent=2' http://127.0.0.1:8080/api/v1/token/
HTTP/1.1 200 OK
Server: gunicorn/19.3.0
Date: Mon, 16 Mar 2015 19:16:42 GMT
Connection: close
Transfer-Encoding: chunked
Vary: Accept, Cookie
Content-Type: application/json; indent=2
Allow: GET, HEAD, OPTIONS

{
  "valid_until": 1426533420,
  "token": 186032,
  "interval": 60,
  "address": "10.0.2.2"
}
```

### 

Once you have the token, authorize the controlling device using the `/api/v1/auth/ADDR` endpoint:

Note the `Ctrl-D`

```sh
$ curl -i -XPOST -d '@-' -H 'Content-Type: application/json; indent=2' http://127.0.0.1:8080/api/v1/auth/10.0.2.2/
{
    "token": 186032
}
Ctrl-D
HTTP 200 OK
Content-Type: application/json
Vary: Accept
Allow: GET, POST, HEAD, OPTIONS

{
    "controlling_ip": "127.0.0.1",
    "controlled_ip": "10.0.2.2"
}
```


### <a name="proxy-security"></a>Proxy Security

If you are using an HTTP proxy such as [nginx](http://nginx.org/), make sure it is configured to set the
`HTTP_X_REAL_IP` header, or token generation will not work.

One security implication of using the `HTTP_X_REAL_IP` field to determine the client address is that the client can
manipulate this field to obtain a token for an arbitrary address. For example, `curl -H 'X_REAL_IP: 1.2.3.4'`.

To prevent this, ATC restricts which clients are allowed to set the `HTTP_X_REAL_IP` request header.
This is done by use of the `PROXY_IPS` field of the `ATC_API` dict in the django settings file:

    ATC_API = {
        'PROXY_IPS': ['1.2.3.4', '2.3.4.5'],
    }
