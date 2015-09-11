/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


 var ShapingButton = React.createClass({
  render: function () {
    button_values = [
      {
        message: "ATC is not running",
        css: "warning",
      },
      {
        message: "Turn Off",
        css: "danger",
      },
      {
        message: "Turn On",
        css: "primary",
      },
      {
        message: "Update Shaping",
        css: "success",
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
      <div>
        <h4>{capitalizeFirstLetter(d) + "link"}:</h4>
        <div className="well" id={d + "_well"}>
          <div className="form-horizontal accordion">
            <CollapseableInputList text="Bandwidth">
              <LinkShapingNumberSetting params={[d, "rate"]} text="Rate" placeholder="in kbps" link_state={this.props.link_state} />
            </CollapseableInputList>
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
      </div>
    );
  }
});


var ShapingSettings = React.createClass({
  render: function () {
    return (
      <div className="panel-group" id="accordion2" role="tablist" aria-multiselectable="false">
        <div className="panel panel-default">
          <div className="panel-heading" data-toggle="collapse" data-parent="#accordion2" href="#collapseShaping" aria-expanded="false" aria-controls="collapseShaping">
              <h4 className="panel-title">
                  Shaping Settings
              </h4>
          </div>
          <div id="collapseShaping" className="panel-collapse collapse in" role="tabpanel">
            <div className="panel-body">
              <div className="row">
                <div className="col-md-6">
                  <LinkShapingSettings direction="up" link_state={this.props.link_state} />
                </div>
                <div className="col-md-6">
                  <LinkShapingSettings direction="down" link_state={this.props.link_state} />
                </div>
              </div>

              <JSONView json={this.props.before} label="Before:" />
              <JSONView json={this.props.after} label="After:" />
            </div>
          </div>
        </div>
      </div>
    );
  }
});