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

var pd = require('./pageData.js');
var cfunc = require('../common/commonFunctions.js');

describe('Traffic Portal Delivery Services Suite', function() {

	const pageData = new pd();
	const commonFunctions = new cfunc();
	const mockVals = {
		dsType: ["ANY MAP", "DNS", "HTTP", "STEERING"],
		active: "Active",
		xmlId: "xml-id-" + commonFunctions.shuffle('abcdefghijklmonpqrstuvwxyz'),
		orgServerFqdn: "http://dstest.com",
		longDesc: "This is only a test that should be disposed of by Automated UI Testing."
	};

	it('should open ds page and click button to create a new one', async () => {
		console.log('Opening delivery services page');
		await browser.setLocation("delivery-services");
		expect(browser.getCurrentUrl().then(commonFunctions.urlPath)).toEqual(commonFunctions.urlPath(browser.baseUrl)+"#!/delivery-services");
	});

	it('should create and select type of ds from the dropdown and confirm', async () => {
		console.log('Clicked Create New and selecting a type');
		await browser.driver.findElement(by.name('createDeliveryServiceButton')).click();
		expect(pageData.selectFormSubmitButton.isEnabled()).toBe(false);
		await browser.driver.findElement(by.name('selectFormDropdown')).sendKeys(mockVals.dsType[1]);
		expect(pageData.selectFormSubmitButton.isEnabled()).toBe(true);
		await pageData.selectFormSubmitButton.click();
		expect(browser.getCurrentUrl().then(commonFunctions.urlPath)).toEqual(commonFunctions.urlPath(browser.baseUrl)+"#!/delivery-services/new?type=" + mockVals.dsType[1]);
	});

	it('should populate and submit the ds form', async () => {
		console.log('Filling out form for ' + mockVals.xmlId);
		expect(pageData.createButton.isEnabled()).toBe(false);
		await pageData.active.click();
		await pageData.active.sendKeys(mockVals.active);
		commonFunctions.selectDropdownbyNum(pageData.type, 1);
		await pageData.xmlId.sendKeys(mockVals.xmlId);
		await pageData.displayName.sendKeys(mockVals.xmlId);
		commonFunctions.selectDropdownbyNum(pageData.tenantId, 1);
		commonFunctions.selectDropdownbyNum(pageData.cdn, 1);
		await pageData.orgServerFqdn.sendKeys(mockVals.orgServerFqdn);
		commonFunctions.selectDropdownbyNum(pageData.protocol, 1);
		await pageData.longDesc.sendKeys(mockVals.longDesc);
		expect(pageData.createButton.isEnabled()).toBe(true);
		await pageData.createButton.click();
		expect(browser.getCurrentUrl().then(commonFunctions.urlPath)).toMatch(commonFunctions.urlPath(browser.baseUrl)+"#!/delivery-services/[0-9]+");
	});

	it('should update the ds', async () => {
		console.log('Updating the form for ' + mockVals.xmlId);
		await browser.setLocation("delivery-services");
		await pageData.searchFilter.sendKeys(mockVals.xmlId);
		await element.all(by.repeater('ds in ::deliveryServices')).filter(function(row){
			return row.element(by.name('xmlId')).getText().then(function(val){
				return val.toString() === mockVals.xmlId.toString();
			});
		}).get(0).click();
		expect(pageData.updateButton.isEnabled()).toBe(false);
		await pageData.displayName.sendKeys(mockVals.xmlId + "updated");
		expect(pageData.updateButton.isEnabled()).toBe(true);
		await pageData.updateButton.click();
		expect(pageData.displayName.getText() === mockVals.name + "updated");
	});

	it('should delete the ds', async () => {
		console.log('Deleting ' + mockVals.xmlId);
		await pageData.deleteButton.click();
		await pageData.confirmWithNameInput.sendKeys(mockVals.xmlId);
		await pageData.deletePermanentlyButton.click();
		expect(browser.getCurrentUrl().then(commonFunctions.urlPath)).toEqual(commonFunctions.urlPath(browser.baseUrl)+"#!/delivery-services");
	});
});
