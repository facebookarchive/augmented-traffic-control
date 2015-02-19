from django.db import models


class Profile(models.Model):
    name = models.CharField(max_length=100, blank=False, null=False)
    content = models.CharField(max_length=1024, blank=False, null=False)
