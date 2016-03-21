<!--
If you are filling an issue, please provide the information below, if this is not a bug and just a question, you can skip it.

---------------------------------------------------
BUG REPORT INFORMATION
---------------------------------------------------
Here are a few questions that will help us get a better understanding of your setup. Please take a minute to answer them, this will avoid some back and forth communication by providing some context to us.
-->
**Meeting the requirements:**
**ATC** has a few requirements, this section is used to make sure the requirements are met. Until you can check all the boxes, **ATC** will not work on your system.
- [ ] `atcd` runs on Linux
- [ ] `atcd` has 2 physical interfaces
 - [ ] 1 that connects to the internet (WAN)
 - [ ] 1 that connects to the local network (LAN)
- [ ] the traffic from the devices that are being shaped is routed through `atcd`
- [ ] `atcd` sees the real IP of the devices (e.g there is no **NAT** on the LAN segment)

**Checking for special atcd options:**

In some cases, `atcd` default configuration may not be suitable for your system. Use:
```
atcd -h
```
do change those default to suit your setup.
For instance, if your lan interface is `wlan0`, you should start `atcd` using:
```
atcd --atcd-lan wlan0
```

**Steps to reproduce the issue:**
1.
2.
3.


**Describe the results you received:**


**Describe the results you expected:**



**Any additional information that you think is important:**


**Dumping system info:**

While this may not be required in many cases, providing the following information will help us better understand your setup.

**NOTE**: before putting the output on gist, please check its content is fine to be publically available

Run https://github.com/facebook/augmented-traffic-control/blob/master/utils/dump_system_info.sh and upload the content of the logs to [https://gist.github.com/](gist)

```
curl https://raw.githubusercontent.com/facebook/augmented-traffic-control/master/utils/dump_system_info.sh | bash -s -
```

This default to using `eth0` and `eth1` for `wan` and `lan` respectively. If you are using other interfaces (let say eth0 and wlan0 respectively), use:
```
curl https://raw.githubusercontent.com/facebook/augmented-traffic-control/master/utils/dump_system_info.sh | bash -s -  eth0 wlan0
```
