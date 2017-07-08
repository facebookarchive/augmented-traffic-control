Setup
========

Requirements
--------

### Linux

`ATC` makes use of [`iproute2`](http://www.linuxfoundation.org/collaborate/workgroups/networking/iproute2) which is only
supported on platforms running a linux kernel.

### Network Gateway

`ATC` is intended to be deployed to a network gateway. Normally this would mean that the machine `ATC` runs on would
require **2 network interfaces**, one for WAN and one for LAN. However it is possible to simulate this setup by making
use of more advanced [virtual interfaces](https://wiki.archlinux.org/index.php/VLAN) and routing options on the host.

### `python 2.7` and `pip`

`ATC` requires `python 2.7` to work correctly, complete with `pip`. 

### virtualenv

Although not strictly required, use of a [virtualenv](https://virtualenv.pypa.io/en/latest/) is recommended.

To setup a new virtualenv in `~/atc/venv`:

```shell
mkdir -p ~/atc
virtualenv ~/atc/venv
source ~/atc/venv/bin/activate
```

On production environments, you probably want to put your atc installation somewhere besides
`~/atc`.

ATC Daemon
--------

Installation:

```shell
sudo pip install atcd
```

Running `atcd` (as root):

```shell
sudo atcd --atcd-lan eth0 --atcd-wan eth1
```

ATC Interface
--------

Install the ATC API and UI packages:

```shell
sudo pip install django-atc-api django-atc-demo-ui django-atc-profile-storage
```

### Django Webapp Setup

ATC's Interfaces are written with [Django](https://www.djangoproject.com/), so they require a
`django` webapp to work correctly.

To create and setup this webapp you will need the `django` python package:

```shell
pip install django
```

To create a new Django webapp:

```shell
cd ~/atc
django-admin startproject atcui
```

Once you have a django webapp setup, you can enable the ATC apps by adding them to the `INSTALLED_APPS` list in
django's `settings.py`:

```python
INSTALLED_APPS = (
    ...
    # Django ATC API
    'rest_framework',
    'atc_api',
    # Django ATC Demo UI
    'bootstrap_themes',
    'django_static_jquery',
    'atc_demo_ui',
    # Django ATC Profile Storage
    'atc_profile_storage',
)
```

Once this is done, you can add the ATC urls to the django webapp's `urls.py`:

```python
from django.views.generic.base import RedirectView

...

urlpatterns = patterns('',
    ...
    # Django ATC API
    url(r'^api/v1/', include('atc_api.urls')),
    # Django ATC Demo UI
    url(r'^atc_demo_ui/', include('atc_demo_ui.urls')),
    # Django ATC profile storage
    url(r'^api/v1/profiles/', include('atc_profile_storage.urls')),
    url(r'^$', RedirectView.as_view(url='/atc_demo_ui/', permanent=False)),
)
```

Migrate the django database:

```shell
python manage.py migrate
```

And finally, run the django server:

```shell
python manage.py runserver 0.0.0.0:8080
```
