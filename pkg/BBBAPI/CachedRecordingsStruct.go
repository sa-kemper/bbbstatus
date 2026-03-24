/*
 * Copyright 2025 Samuel Kemper
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package BBBAPI

import (
	db "bbbstatus/internal/database"
	"context"
	"fmt"
	"log"
	"time"
)

type CachedRecordings struct {
	lastRequestTime         time.Time
	cacheInvalidationPeriod time.Duration
	items                   []*Recording
	api                     *API
}

func (c *CachedRecordings) GetAll() (allItems []interface{}, err error) {
	if c.lastRequestTime.After(time.Now().Add(c.cacheInvalidationPeriod)) {
		err = c.Update(context.TODO())
		if err != nil {
			log.Fatal("Error updating cached recordings, cache is invalid.")
			return allItems, err
		}
	}
	for _, item := range c.items {
		allItems = append(allItems, *item)
	}
	return
}

func (c *CachedRecordings) GetByFilter(filter func(obj interface{}) bool) (filteredObjects []interface{}, err error) {
	if c.lastRequestTime.After(time.Now().Add(c.cacheInvalidationPeriod)) {
		err = c.Update(context.TODO())
		if err != nil {
			log.Fatal("Error updating cached recordings, cache is invalid.")
			return filteredObjects, err
		}
	}

	for _, item := range c.items {
		if !filter(item) {
			continue
		}

		filteredObjects = append(filteredObjects, *item)
	}
	return
}

func (c *CachedRecordings) Update(ctx context.Context) error {
	recordings, err := c.api.GetRecordings(ctx, GetRecordingsParameters{}, db.Meeting{})
	if err != nil {
		fmt.Println("error updating cached recordings")
		return err
	}

	c.lastRequestTime = time.Now()
	recordingsPointers := make([]*Recording, len(recordings.Recording))
	for rpi := range recordingsPointers {
		recordingsPointers[rpi] = &recordings.Recording[rpi]
	}
	c.items = recordingsPointers
	return nil
}

func (c *CachedRecordings) GetFilteredRecordings(ctx context.Context, params GetRecordingsParameters, meeting db.Meeting) (result Recordings, err error) {
	return c.api.GetRecordings(ctx, params, meeting)
}
