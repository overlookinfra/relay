/*
 * Relay API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: v20200615
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// NotificationAllOf An external event
type NotificationAllOf struct {
	// The current user has marked this notification done
	Done *bool `json:"done,omitempty"`
	// The fields to use for linking out from the notification
	Fields *map[string]interface{} `json:"fields,omitempty"`
	// Whether the current user has read this notification
	Read bool `json:"read"`
	// The state of this notification
	State string `json:"state"`
}

// NewNotificationAllOf instantiates a new NotificationAllOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNotificationAllOf(read bool, state string) *NotificationAllOf {
	this := NotificationAllOf{}
	this.Read = read
	this.State = state
	return &this
}

// NewNotificationAllOfWithDefaults instantiates a new NotificationAllOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNotificationAllOfWithDefaults() *NotificationAllOf {
	this := NotificationAllOf{}
	return &this
}

// GetDone returns the Done field value if set, zero value otherwise.
func (o *NotificationAllOf) GetDone() bool {
	if o == nil || o.Done == nil {
		var ret bool
		return ret
	}
	return *o.Done
}

// GetDoneOk returns a tuple with the Done field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NotificationAllOf) GetDoneOk() (*bool, bool) {
	if o == nil || o.Done == nil {
		return nil, false
	}
	return o.Done, true
}

// HasDone returns a boolean if a field has been set.
func (o *NotificationAllOf) HasDone() bool {
	if o != nil && o.Done != nil {
		return true
	}

	return false
}

// SetDone gets a reference to the given bool and assigns it to the Done field.
func (o *NotificationAllOf) SetDone(v bool) {
	o.Done = &v
}

// GetFields returns the Fields field value if set, zero value otherwise.
func (o *NotificationAllOf) GetFields() map[string]interface{} {
	if o == nil || o.Fields == nil {
		var ret map[string]interface{}
		return ret
	}
	return *o.Fields
}

// GetFieldsOk returns a tuple with the Fields field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NotificationAllOf) GetFieldsOk() (*map[string]interface{}, bool) {
	if o == nil || o.Fields == nil {
		return nil, false
	}
	return o.Fields, true
}

// HasFields returns a boolean if a field has been set.
func (o *NotificationAllOf) HasFields() bool {
	if o != nil && o.Fields != nil {
		return true
	}

	return false
}

// SetFields gets a reference to the given map[string]interface{} and assigns it to the Fields field.
func (o *NotificationAllOf) SetFields(v map[string]interface{}) {
	o.Fields = &v
}

// GetRead returns the Read field value
func (o *NotificationAllOf) GetRead() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Read
}

// GetReadOk returns a tuple with the Read field value
// and a boolean to check if the value has been set.
func (o *NotificationAllOf) GetReadOk() (*bool, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Read, true
}

// SetRead sets field value
func (o *NotificationAllOf) SetRead(v bool) {
	o.Read = v
}

// GetState returns the State field value
func (o *NotificationAllOf) GetState() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.State
}

// GetStateOk returns a tuple with the State field value
// and a boolean to check if the value has been set.
func (o *NotificationAllOf) GetStateOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.State, true
}

// SetState sets field value
func (o *NotificationAllOf) SetState(v string) {
	o.State = v
}

func (o NotificationAllOf) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Done != nil {
		toSerialize["done"] = o.Done
	}
	if o.Fields != nil {
		toSerialize["fields"] = o.Fields
	}
	if true {
		toSerialize["read"] = o.Read
	}
	if true {
		toSerialize["state"] = o.State
	}
	return json.Marshal(toSerialize)
}

type NullableNotificationAllOf struct {
	value *NotificationAllOf
	isSet bool
}

func (v NullableNotificationAllOf) Get() *NotificationAllOf {
	return v.value
}

func (v *NullableNotificationAllOf) Set(val *NotificationAllOf) {
	v.value = val
	v.isSet = true
}

func (v NullableNotificationAllOf) IsSet() bool {
	return v.isSet
}

func (v *NullableNotificationAllOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNotificationAllOf(val *NotificationAllOf) *NullableNotificationAllOf {
	return &NullableNotificationAllOf{value: val, isSet: true}
}

func (v NullableNotificationAllOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNotificationAllOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

