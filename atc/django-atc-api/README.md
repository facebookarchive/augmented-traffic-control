=======
ATC Api
=======

ATC API is a Django app that allow to bridge a REST API to ATCD's thrift API.

Quick start
-----------

1. Add "atc_api" to your INSTALLED_APPS setting like this::

    INSTALLED_APPS = (
        ...
        'atc_api',
        'rest_framework',
    )

2. Include the atc_api URLconf in your project urls.py like this::

    url(r'^api/v1/', include('atc_api.urls')),

3. Start the development server

4. Visit http://127.0.0.1:8000/api/v1/shape/ to set/unset shaping.
