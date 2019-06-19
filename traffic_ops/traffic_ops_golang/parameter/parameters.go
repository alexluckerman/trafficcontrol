package parameter

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/lib/go-tc/tovalidate"
	"github.com/apache/trafficcontrol/lib/go-util"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/auth"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/dbhelpers"

	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	NameQueryParam       = "name"
	SecureQueryParam     = "secure"
	ConfigFileQueryParam = "configFile"
	IDQueryParam         = "id"
	ValueQueryParam      = "value"
)

var (
	HiddenField = "********"
)

//we need a type alias to define functions on
type TOParameter struct {
	api.APIInfoImpl `json:"-"`
	tc.ParameterNullable
}

func (v *TOParameter) SetLastUpdated(t tc.TimeNoMod) { v.LastUpdated = &t }
func (v *TOParameter) InsertQuery() string           { return insertQuery() }
func (v *TOParameter) NewReadObj() interface{}       { return &tc.ParameterNullable{} }
func (v *TOParameter) SelectQuery() string           { return selectQuery() }
func (v *TOParameter) ParamColumns() map[string]dbhelpers.WhereColumnInfo {
	return map[string]dbhelpers.WhereColumnInfo{
		ConfigFileQueryParam: dbhelpers.WhereColumnInfo{"p.config_file", nil},
		IDQueryParam:         dbhelpers.WhereColumnInfo{"p.id", api.IsInt},
		NameQueryParam:       dbhelpers.WhereColumnInfo{"p.name", nil},
		SecureQueryParam:     dbhelpers.WhereColumnInfo{"p.secure", api.IsBool}}
}
func (v *TOParameter) UpdateQuery() string { return updateQuery() }
func (v *TOParameter) DeleteQuery() string { return deleteQuery() }

func (param TOParameter) GetKeyFieldsInfo() []api.KeyFieldInfo {
	return []api.KeyFieldInfo{{IDQueryParam, api.GetIntKey}}
}

//Implementation of the Identifier, Validator interface functions
func (param TOParameter) GetKeys() (map[string]interface{}, bool) {
	if param.ID == nil {
		return map[string]interface{}{IDQueryParam: 0}, false
	}
	return map[string]interface{}{IDQueryParam: *param.ID}, true
}

func (param *TOParameter) SetKeys(keys map[string]interface{}) {
	i, _ := keys[IDQueryParam].(int) //this utilizes the non panicking type assertion, if the thrown away ok variable is false i will be the zero of the type, 0 here.
	param.ID = &i
}

func (param *TOParameter) GetAuditName() string {
	if param.Name != nil {
		return *param.Name
	}
	if param.ID != nil {
		return strconv.Itoa(*param.ID)
	}
	return "unknown"
}

func (param *TOParameter) GetType() string {
	return "param"
}

// Validate fulfills the api.Validator interface
func (param TOParameter) Validate() error {
	// Test
	// - Secure Flag is always set to either 1/0
	// - Admin rights only
	// - Do not allow duplicate parameters by name+config_file+value
	// - Client can send NOT NULL constraint on 'value' so removed it's validation as .Required
	errs := validation.Errors{
		NameQueryParam:       validation.Validate(param.Name, validation.Required),
		ConfigFileQueryParam: validation.Validate(param.ConfigFile, validation.Required),
	}

	return util.JoinErrs(tovalidate.ToErrors(errs))
}

func (pa *TOParameter) Create() (error, error, int) {
	if pa.Value == nil {
		pa.Value = util.StrPtr("")
	}
	return api.GenericCreate(pa)
}

func (param *TOParameter) Read() ([]interface{}, error, error, int) {
	queryParamsToQueryCols := param.ParamColumns()
	where, orderBy, limit, queryValues, errs := dbhelpers.BuildWhereAndOrderByAndPagination(param.APIInfo().Params, queryParamsToQueryCols)
	if len(errs) > 0 {
		return nil, util.JoinErrs(errs), nil, http.StatusBadRequest
	}

	query := selectQuery() + where + ParametersGroupBy() + orderBy + limit
	rows, err := param.ReqInfo.Tx.NamedQuery(query, queryValues)
	if err != nil {
		return nil, nil, errors.New("querying " + param.GetType() + ": " + err.Error()), http.StatusInternalServerError
	}
	defer rows.Close()

	params := []interface{}{}
	for rows.Next() {
		var p tc.ParameterNullable
		if err = rows.StructScan(&p); err != nil {
			return nil, nil, errors.New("scanning " + param.GetType() + ": " + err.Error()), http.StatusInternalServerError
		}
		if p.Secure != nil && *p.Secure && param.ReqInfo.User.PrivLevel < auth.PrivLevelAdmin {
			p.Value = &HiddenField
		}
		params = append(params, p)
	}

	return params, nil, nil, http.StatusOK
}

func (pa *TOParameter) Update() (error, error, int) {
	if pa.Value == nil {
		pa.Value = util.StrPtr("")
	}
	return api.GenericUpdate(pa)
}

func (pa *TOParameter) Delete() (error, error, int) { return api.GenericDelete(pa) }

func insertQuery() string {
	query := `INSERT INTO parameter (
name,
config_file,
value,
secure) VALUES (
:name,
:config_file,
:value,
:secure) RETURNING id,last_updated`
	return query
}

func selectQuery() string {

	query := `SELECT
p.config_file,
p.id,
p.last_updated,
p.name,
p.value,
p.secure,
COALESCE(array_to_json(array_agg(pr.name) FILTER (WHERE pr.name IS NOT NULL)), '[]') AS profiles
FROM parameter p
LEFT JOIN profile_parameter pp ON p.id = pp.parameter
LEFT JOIN profile pr ON pp.profile = pr.id`
	return query
}

func updateQuery() string {
	query := `UPDATE
parameter SET
config_file=:config_file,
id=:id,
name=:name,
value=:value,
secure=:secure
WHERE id=:id RETURNING last_updated`
	return query
}

// ParametersGroupBy ...
func ParametersGroupBy() string {
	groupBy := ` GROUP BY p.config_file, p.id, p.last_updated, p.name, p.value, p.secure`
	return groupBy
}

func deleteQuery() string {
	query := `DELETE FROM parameter
WHERE id=:id`
	return query
}
