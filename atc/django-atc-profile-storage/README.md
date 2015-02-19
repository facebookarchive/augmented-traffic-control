===================
ATC Profile Storage
===================

ATC Profile Storage is a Django app that allows to storing predefined ATC
profiles in DB.

Quick start
-----------

1. Add "atc_profile_storage" to your INSTALLED_APPS setting like this::

    INSTALLED_APPS = (
        ...
        'atc_profile_storage',
        'rest_framework',
    )

2. Include the atc_profile_storage URLconf in your project urls.py like this::

    url(r'^api/v1/profile', include('atc_profile_storage.urls')),

3. Start the development server

4. Visit http://127.0.0.1:8000/api/v1/profile .
