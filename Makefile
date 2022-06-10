SIDELINE_PLUGIN_DIRS=$(wildcard ./plugins/sideline/*)

all: build-plugins

clean: clean-plugins

#build-plugins: $(GREETER_PLUGIN_DIRS) $(CLUBBER_PLUGIN_DIRS) $(DUBBER_PLUGIN_DIRS) $(SIDELINE_PLUGIN_DIRS)
build-plugins: $(SIDELINE_PLUGIN_DIRS)

clean-plugins: 
	rm -f ./plugins/built/*

$(SIDELINE_PLUGIN_DIRS):
	$(info Sideline plugins at: $(SIDELINE_PLUGIN_DIRS))
	$(MAKE) -C $@

#.PHONY: all $(GREETER_PLUGIN_DIRS) $(CLUBBER_PLUGIN_DIRS) $(DUBBER_PLUGIN_DIRS) $(SIDELINE_PLUGIN_DIRS)
.PHONY: all $(SIDELINE_PLUGIN_DIRS)