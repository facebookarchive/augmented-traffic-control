from django.conf.urls import patterns, url
from atc_profile_storage import views

urlpatterns = [
    url(r'^$', views.profile_list),
    url(r'^(?P<pk>[0-9]+)/$', views.profile_detail),
]
