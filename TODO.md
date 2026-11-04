# Things to do

* Lock should be driven by daemon, not by individual operations.

* Random and Sequential need to be per calendar entry, not overall (or maybe both)

* Scene overides not daily but by time period. 
* Daily overrides additive, not replaceive (or maybe both)
* Responses updated to list overrided scenes along with their time periods.

* MQTT channels and REST ops to show available scenes, colors etc.. For MQTT, messages should be retained.

* Resolve import loop woes, maybe channels in RPC to drive all operations towards daemon

* Define patriotic scene to use for the next few days

* Light message frequency seems excessive, only publish on changes, ensure retained

