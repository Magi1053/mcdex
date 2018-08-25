// ***************************************************************************
//
//  Copyright 2017 David (Dizzy) Smith, dizzyd@dizzyd.com
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
// ***************************************************************************

package main

import (
	"fmt"
	"strconv"

	"github.com/Jeffail/gabs"
	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

// CurseForgeFile is struct which describe a file on CurseForge
type CurseForgeFile struct {
	ID        int
	ProjectID int
	Desc      string
}

func getCurseForgeFile(url string) (CurseForgeFile, error) {
	var result CurseForgeFile

	// Retrieve the URL (we assume it's a HTML webpage)
	res, e := HttpGet(url)
	if e != nil {
		return result, fmt.Errorf("failed to get %s: %+v", url, e)
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return result, fmt.Errorf("failed to parse %s: %+v", url, e)
	}

	// Setup a JS VM and run the HTML through it; we want to process any
	// script sections in the head so we can extract Elerium meta-data
	vm := otto.New()
	vm.Run("Elerium = {}; Elerium.ProjectFileDetails = {}")
	doc.Find("head script").Each(func(i int, sel *goquery.Selection) {
		vm.Run(sel.Text())
	})

	// Convert the Elerium data into JSON, then a string to get it out the VM
	data, err := vm.Run("JSON.stringify(Elerium.ProjectFileDetails)")
	if err != nil {
		return result, fmt.Errorf("failed to extract project file details: %+v", err)
	}

	// Reparse from string into JSON (blech)
	dataStr, _ := data.ToString()
	projectDetails, _ := gabs.ParseJSON([]byte(dataStr))

	// Store all the data into the result and return
	result.Desc, _ = doc.Find("meta[property='og:description']").Attr("content")
	result.ProjectID, _ = strconv.Atoi(projectDetails.S("projectID").Data().(string))
	result.ID, _ = strconv.Atoi(projectDetails.S("projectFileID").Data().(string))
	return result, nil
}
