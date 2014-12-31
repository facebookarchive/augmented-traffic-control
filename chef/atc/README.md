# atc cookbook

Install and configure atc

# Requirements
## Platforms
- Centos 6
- Ubuntu

## Cookbooks
- python
- yum-epel
- sysctl

NOTE: The `yum-epel` is only used on CentOS

# Usage

python recipe must be included before any atc related ones.

If you want to install both `atcd` and `atcui` include the default `atc` recipe in your run list
If you only want the `atcd` daemon, include `atc::atcd` in your run list
For the `atcui` only, include `atc::atcui` in your run list

# Attributes

# Recipes

## atc::default
Install both the daemon and the ui
## atc::atcd
Install the daemon only
## atc::atcui
Install the UI only

# Author

Author:: Emmanuel Bretelle (<chantra@fb.com>)
