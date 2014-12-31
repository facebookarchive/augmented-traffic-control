===========
ATC DEMO UI
===========

Django ATC Demo UI is a Django app that allow to modify traffic shaping applied
to a device via a Web UI.

Quick start
-----------

1. Add "atc_demo_ui" to your INSTALLED_APPS setting like this::

    INSTALLED_APPS = (
        ...
        'atc_demo_ui',
    )

2. Include the atc URLconf in your project urls.py like this::

    url(r'^atc_demo/', include('atc_demo.urls')),

3. Start the development server

4. Visit http://127.0.0.1:8000/atc_demo to access ATC Demo UI.
