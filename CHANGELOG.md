# Changelog

* 0.1.3
    * Support Django rest framework 3.2

* 0.1.2
    * Fixes:
        * Better detection of logging subsystem and don't fail if /dev/log does not exist
        * Regenerating thrift files was overriding package version #107
        * Fixes the UI to run in older browsers
    * Featues:
        * UI improvements
        * Added remote controlling functionality to the UI
        * Added a set of sample profiles #56
        * Created a dockerized instance of ATC
        * Added --atcd-dont-drop-packets flag to not drop packets when going beyond the max bandwidth. Packets will be buffered instead.
    * Misc:
        * More unittest
        * More documentation fixes
        * Shape test util
        * Script to dump system config for troubleshooting purpose
        * Code refactoring
        * Updated to React 0.13.3

* 0.1.1
    * Fixes:
        * Fix profile creation in Firefox #59
        * Fix installing packages through pip when wheel is not used #77
    * Misc:
        * Added some unittest
        * Build and test PR on Travis
        * Bunch of typo fixes
        * More documentation

* 0.1.0
    * Initial Release
