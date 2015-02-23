from atc_profile_storage.models import Profile
from rest_framework import serializers
import ast
import json



class ProfileSerializer(serializers.ModelSerializer):
    class Meta:
        model = Profile
        fields = ('id', 'name', 'content')

    def to_representation(self, instance):
        ret = super(serializers.ModelSerializer, self).to_representation(instance)
        ret['content'] = ast.literal_eval(ret['content'])
        return ret
