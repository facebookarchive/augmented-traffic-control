# ATC Profile Storage

ATC Profile Storage is a Django app that allows to store predefined ATC profiles in DB.

## Requirements

* [Django 1.7](https://github.com/django/django)
* [atc_api](../django-atc-api)

`ATC Profile Storage` depends on `ATC API` so make sure you have installed and configured [ATC API](../django-atc-api) first.

## Installation

The easiest way to install `django-atc-profile-storage` is to install it directly from [pip](https://pypi.python.org/pypi).

### From pip
```bash
pip install django-atc-profile-storage
```

### From source
```bash
cd path/to/django-atc-profile-storage
pip install .
```

## Configuration

1. Add `atc_profile_storage` to your `INSTALLED_APPS`' `settings.py` like this:
```python
    INSTALLED_APPS = (
        ...
        'atc_profile_storage',
        'rest_framework',
    )
```

2. Include the `atc_profile_storage` URLconf in your project `urls.py` like this:
```python
    ...
    url(r'^api/v1/profiles/', include('atc_profile_storage.urls')),
    ...
```

3. Migrate the Django DB:
```bash
python manage.py migrate
```

4. Start the development server
```bash
python manage.py runserver 0.0.0.0:8000
```

5. Visit http://127.0.0.1:8000/api/v1/profiles .
