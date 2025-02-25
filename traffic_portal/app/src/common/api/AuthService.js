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

var AuthService = function($rootScope, $http, $state, $location, userModel, messageModel, ENV) {

    this.login = function(username, password) {
        userModel.resetUser();
        return $http.post(ENV.api['root'] + 'user/login', { u: username, p: password }).then(
            function(result) {
                $rootScope.$broadcast('authService::login');
                const redirect = decodeURIComponent($location.search().redirect);
                if (redirect !== 'undefined') {
                    $location.search('redirect', null); // remove the redirect query param
                    $location.url(redirect);
                } else {
                    $location.url('/');
                }
                return result;
            },
            function(err) {
                throw err;
            }
        );
    };

    this.tokenLogin = function(token) {
        userModel.resetUser();
        return $http.post(ENV.api['root'] + "user/login/token", { t: token }).then(
            function(result) {
                return result;
            },
            function(err) {
                throw err;
            }
        );
    };

    this.logout = function() {
        userModel.resetUser();
        return $http.post(ENV.api['root'] + 'user/logout').then(
            function(result) {
                $rootScope.$broadcast('trafficPortal::exit');
                if ($state.current.name == 'trafficPortal.public.login') {
                    messageModel.setMessages(result.alerts, false);
                } else {
                    messageModel.setMessages(result.alerts, true);
                    $state.go('trafficPortal.public.login');
                }
                return result;
            },
            function(err) {
                throw err;
            }
        );
    };

};

AuthService.$inject = ['$rootScope', '$http', '$state', '$location', 'userModel', 'messageModel', 'ENV'];
module.exports = AuthService;
