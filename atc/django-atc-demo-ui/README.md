# ATC DEMO UI

Django ATC Demo UI is a Django app that allow to modify traffic shaping applied to a device via a Web UI.

Even though it is a Django app, `ATC Demo UI` is mostly a [React](http://facebook.github.io/react/) application that uses [Bootstrap](http://getbootstrap.com/) to make the app responsive.

## Requirements

* [Django 1.7](https://github.com/django/django)
* [atc_api](../django-atc-api)

`ATC Demo UI` depends on `ATC API` so make sure you have installed and configured the [ATC API](../django-atc-api) first.

## Installation

The easiest way to install `django-atc-demo-ui` is to install it directly from [pip](https://pypi.python.org/pypi).

### From pip
```bash
pip install django-atc-demo-ui
```

### From source
```bash
cd path/to/django-atc-demo-ui
pip install .
```

## Configuration

1. Add `atc_demo_ui` and its dependencies to your `INSTALLED_APPS`' `settings.py` like this:
```python
    INSTALLED_APPS = (
        ...
        'bootstrap_themes',
        'django_static_jquery',
        'atc_demo_ui',
    )
```
2. Include the `atc_demo_ui` URLconf in your project `urls.py` like this:

    url(r'^atc_demo_ui/', include('atc_demo_ui.urls')),

If you want to have `/` redirecting to `/atc/demo_ui`, you can update `urls.py`
```python
...
from django.views.generic.base import RedirectView

urlpatterns = patterns('',
    ...
    ...
    url(r'^atc_demo_ui/', include('atc_demo_ui.urls')),
    url(r'^$', RedirectView.as_view(url='/atc_demo_ui/', permanent=False)),
)
```

3. Start the development server
```bash
python manage.py runserver 0.0.0.0:8000
```

4. Visit http://127.0.0.1:8000/atc_demo_ui to access ATC Demo UI.


Some settings like the REST endpoint can be changed in your Dkango project'settings.py:

```python
ATC_DEMO_UI = {
    'REST_ENDPOINT': '/api/v1/',
}
```

see [ATC Demo UI settings](atc_demo_ui/settings.py) for more details.

