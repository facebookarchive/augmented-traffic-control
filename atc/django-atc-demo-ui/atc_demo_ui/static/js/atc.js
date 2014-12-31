/** @jsx React.DOM */

/** https://gist.github.com/NV/8622188 **/
/**
 * RecursiveLinkStateMixin is a LinkState alternative that can update keys in
 * a dictionnary recursively.
 * You can either give it a string of keys separated by a underscore (_)
 * or a list of keys
 */
var RecursiveLinkStateMixin = {
    linkState: function (path) {
        function setPath (obj, path, value) {
            var leaf = resolvePath(obj, path);
            leaf.obj[leaf.key] = value;
        }

        function getPath (obj, path) {
            var leaf = resolvePath(obj, path);
            return leaf.obj[leaf.key];
        }

        function resolvePath (obj, keys) {
            if (typeof keys === 'string') {
                keys = keys.split('_');
            }
            var lastIndex = keys.length - 1;
            var current = obj;
            for (var i = 0; i < lastIndex; i++) {
                var key = keys[i];
                current = current[key];
            }
            return {
                obj: current,
                key: keys[lastIndex]
            };
        }

        return {
            value: getPath(this.state, path),
            requestChange: function(newValue) {
                setPath(this.state, path, newValue);
                this.forceUpdate();
            }.bind(this)
        };
    }
};

var IdentifyableObject = {
    getIdentifier: function () {
        return this.props.params.join('_');
    },
};

var atc_status = {
    OFFLINE: 0,
    ACTIVE: 1,
    INACTIVE: 2,
    OUTDATED: 3,
};


var ShapingButton = React.createClass({
    render: function () {
        button_values = [
            {
                message: "ATC is not running",
                css: "danger",
            },
            {
                message: "Turn Off",
                css: "success",
            },
            {
                message: "Turn On",
                css: "default",
            },
            {
                message: "Update Shaping",
                css: "warning",
            },
        ];

        content = button_values[this.props.status];
        return (
            <button type="button" id={this.props.id} className={"btn btn-" + content.css} disabled={this.props.status == atc_status.OFFLINE} onClick={this.props.onClick}>
                {content.message}
            </button>
        );
    }
});


var LinkShapingNumberSetting = React.createClass({
    mixins: [IdentifyableObject],
    render: function () {
        id = this.getIdentifier();
        link_state = this.props.link_state("settings_" + id);
        return (
            <div className="form-group">
                <label htmlFor={id} className="col-sm-3 control-label">{this.props.text}</label>
                <div className="col-sm-9">
                    <input type="number" defaultValue={link_state.value} className="form-control" id={id} placeholder={this.props.placeholder} min="0" max={this.props.max_value} valueLink={link_state} />
                </div>
            </div>
        )
    }
});


var LinkShapingPercentSetting = React.createClass({
    render: function () {
        return (
            <LinkShapingNumberSetting input_id={this.props.input_id} text={this.props.text} placeholder="In %" link_state={this.props.link_state} max_value="100" />
        )
    }
});


var CollapseableInputList = React.createClass({
    render: function () {
        return (
            <fieldset className="accordion-group">
                <legend>{this.props.text}</legend>
                {this.props.children}
            </fieldset>
        );
    }
});


var CollapseableInputGroup = React.createClass({
    mixins: [IdentifyableObject],
    getInitialState: function () {
        return {collapsed: true};
    },

    handleClick: function (e) {
        this.setState({collapsed: !this.state.collapsed})
    },

    render: function () {
        id = this.getIdentifier();
        var text = this.state.collapsed ? 'Show more' : 'Show less';
        return (
            <div>
                <div className="accordion-heading">
                    <a className="accordion-toggle" data-toggle="collapse" data-target={"#" + id} href="#" onClick={this.handleClick}>{text}</a>
                </div>
                <div className="accordion-body collapse" id={id}>
                    <div className="accordion-inner">
                        {this.props.children}
                    </div>
                </div>
            </div>
        );
    }
});


var LinkShapingSettings = React.createClass({
    render: function () {
        d = this.props.direction;
        return (
            <div className="well" id={d + "_well"}>
                <h3>{d + "link"}</h3>
                <div className="form-horizontal accordion">
                    <LinkShapingNumberSetting params={[d, "rate"]} text="Bandwidth" placeholder="in kbps" link_state={this.props.link_state} />
                    <CollapseableInputList text="Latency">
                        <LinkShapingNumberSetting params={[d, "delay", "delay"]} text="Delay" placeholder="in ms" link_state={this.props.link_state} />
                        <CollapseableInputGroup params={[d, "delay", "collapse"]}>
                            <LinkShapingNumberSetting params={[d, "delay","jitter"]} text="Jitter" placeholder="in %" link_state={this.props.link_state} />
                            <LinkShapingNumberSetting params={[d, "delay", "correlation"]} text="Correlation" placeholder="in %" link_state={this.props.link_state} />
                        </CollapseableInputGroup>
                    </CollapseableInputList>
                    <CollapseableInputList text="Loss">
                        <LinkShapingNumberSetting params={[d, "loss", "percentage"]} text="Percentage" placeholder="in %" link_state={this.props.link_state} />
                        <CollapseableInputGroup params={[d, "loss", "collapse"]}>
                            <LinkShapingNumberSetting params={[d, "loss", "correlation"]} text="Correlation" placeholder="in %" link_state={this.props.link_state} />
                        </CollapseableInputGroup>
                    </CollapseableInputList>
                    <CollapseableInputList text="Corruption">
                        <LinkShapingNumberSetting params={[d, "corruption", "percentage"]} text="Percentage" placeholder="in %" link_state={this.props.link_state} />
                        <CollapseableInputGroup params={[d, "corruption", "collapse"]}>
                            <LinkShapingNumberSetting params={[d, "corruption", "correlation"]} text="Correlation" placeholder="in %" link_state={this.props.link_state} />
                        </CollapseableInputGroup>
                    </CollapseableInputList>
                    <CollapseableInputList text="Reorder">
                        <LinkShapingNumberSetting params={[d, "reorder", "percentage"]} text="Percentage" placeholder="in %" link_state={this.props.link_state} />
                        <CollapseableInputGroup params={[d, "reorder", "collapse"]}>
                            <LinkShapingNumberSetting params={[d, "reorder", "correlation"]} text="Correlation" placeholder="in %" link_state={this.props.link_state} />
                            <LinkShapingNumberSetting params={[d, "reorder", "gap"]} text="Gap" placeholder="integer" link_state={this.props.link_state}/>
                        </CollapseableInputGroup>
                    </CollapseableInputList>
                </div>
            </div>
        );
    }
});


var ShapingSettings = React.createClass({
    render: function () {
        return (
            <div>
            <div className="col-md-6">
                <LinkShapingSettings direction="up" link_state={this.props.link_state} />
            </div>
            <div className="col-md-6">
                <LinkShapingSettings direction="down" link_state={this.props.link_state} />
            </div>
            </div>
        );
    }
});


var ErrorBox = React.createClass({
    render: function () {
        var errors = this.props.error.detail;
        if (typeof this.props.error.detail === 'string') {
            errors = Array(this.props.error.detail);
        } else if (typeof this.props.error.detail === 'object') {
            errors = this.props.error.detail;
        }
        var errorNodes = errors.map(function(error) {
            return (
                <li>{error}</li>
            );
        });
        return (
            <div className="alert alert-danger" role="alert">
                <ul>
                {errorNodes}
                </ul>
            </div>
        )
    }
});


var Atc = React.createClass({
    mixins: [RecursiveLinkStateMixin],
    getInitialState: function() {
        return {
            client: new AtcRestClient(this.props.endpoint),
            settings: new AtcSettings().getDefaultSettings(),
            current_settings: new AtcSettings().getDefaultSettings(),
            status: atc_status.OFFLINE,
            error_msg: "",
        };
    },

    componentDidMount: function() {
        this.getCurrentShaping();
        /** FIXME we are calling getCurrentShaping to make sure that
         * current_settings === settings.... let's be smarter than that.
         */
        this.getCurrentShaping();
    },

    handleClick: function(e) {
        if (e.type == "click") {
            if (this.state.status == atc_status.ACTIVE) {
                this.unsetShaping();
            } else if (this.state.status == atc_status.INACTIVE) {
                this.setShaping();
            }
        }
    },

    updateClick: function(e) {
        if (e.type == "click") {
            this.setShaping();
        }
    },

    hasChanged: function() {
        /* TODO: improve object comparaison e.g null == "", 1 == "1"*/
        function objectEquals(x, y) {
            if (typeof(x) === 'number') {
                x = x.toString();
            }
            if (typeof(y) === 'number') {
                y = y.toString();
            }
            if (typeof(x) != typeof(y)) {
                return false;
            }

            if (Array.isArray(x) || Array.isArray(y)) {
                return x.toString() === y.toString();
            }

            if (x === null && y === null) {
                return true;
            }

            if (typeof(x) === 'object' && x !== null) {
                x_keys = Object.keys(x);
                y_keys = Object.keys(y);
                if (x_keys.sort().toString() !== y_keys.sort().toString()) {
                    console.error('Object do not have the same keys: ' +
                        x_keys.sort().toString() + ' vs ' +
                        y_keys.sort().toString()
                    );
                    return false;
                }
                equals = true;
                for (key of x_keys) {
                    equals &= objectEquals(x[key], y[key]);
                }
                return equals;
            }
            return x.toString() === y.toString();
        }
        return !objectEquals(this.state.settings, this.state.current_settings);
    },

    getCurrentShaping: function() {
        console.log('getCurrentShaping');
        this.state.client.getCurrentShaping(function (result) {
            if (result.status == 404) {
                this.setState({
                    status: atc_status.INACTIVE,
                    error_msg: '',
                    settings: new AtcSettings().getDefaultSettings(),
                    current_settings: new AtcSettings().getDefaultSettings(),
                });
            } else if (result.status >= 200 && result.status < 300) {
                this.setState({
                    status: atc_status.ACTIVE,
                    error_msg: '',
                    settings: result.json,
                    current_settings: this.state.settings,
                });
            } else {
                this.setState({
                    status: atc_status.OFFLINE,
                    error_msg: result.json,
                    settings: new AtcSettings().getDefaultSettings(),
                });
            }
        }.bind(this));
    },

    unsetShaping: function() {
        console.log('unsetShaping');
        this.state.client.unshape(function (result) {
            if (result.status >= 200 && result.status < 300) {
                this.setState({
                    status: atc_status.INACTIVE,
                    settings: new AtcSettings().getDefaultSettings(),
                    current_settings: new AtcSettings().getDefaultSettings(),
                });
            } else if (result.status >= 500) {
                this.setState({
                    status: atc_status.OFFLINE,
                    error_msg: result.json,
                });
            }
        }.bind(this));
    },


    setShaping: function() {
        console.log('setShaping');
        this.state.client.shape(function (result) {
            if (result.status >= 200 && result.status < 300) {
                this.setState({
                    status: atc_status.ACTIVE,
                    error_msg: '',
                    settings: result.json,
                    current_settings: {down: this.state.settings.down, up: this.state.settings.up},
                });
            } else if (result.status == 400) {
                var errors = Array();
                for (var key in result.json) {
                    errors = errors.concat(result.data[key].map(function(msg) {
                        return key + ': ' + msg;
                    }));
                }
                this.setState({
                    error_msg: errors,
                });
            } else if (result.status >= 500) {
                this.setState({
                    status: atc_status.OFFLINE,
                    error_msg: result.json,
                });
            }

        }.bind(this), {down: this.state.settings.down, up: this.state.settings.up});
    },

    render: function () {
        link_state = this.linkState;
        var err_msg = "";
        var update_button = "";
        if (this.state.error_msg != "") {
            err_msg = <ErrorBox error={this.state.error_msg} />
        }
        if (this.hasChanged()) {
            update_button = <ShapingButton id="update_button" status={atc_status.OUTDATED} onClick={this.updateClick} />
        }
        return (
            <div>
            <div className="row">
                <div id="shaping_buttons" className="col-md-12 text-center">
                    {update_button}
                    <ShapingButton id="shaping_button" status={this.state.status} onClick={this.handleClick} />
                    {err_msg}
                </div>
            </div>
            <div className="row">
                <ShapingSettings link_state={link_state} />
            </div>
            <div className="row">
                <pre className="col-md-6">{ JSON.stringify(this.state.settings) }</pre>
                <pre className="col-md-6">{ JSON.stringify(this.state.current_settings) }</pre>
            </div>
            </div>
        )
    }
});
