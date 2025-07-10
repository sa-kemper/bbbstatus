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

package BBBEvents

import (
	db "bbbstatus/internal/database"
	"context"
	"fmt"
)

func loadAdditionalUserData(ctx context.Context, user *User, dbQueries *db.Queries) (err error) {
	if user == nil {
		fmt.Println("error occured whilst loadAdditionalUserData, user is nil.")
		return fmt.Errorf("user is nil")
	}
	if user.InternalUserID == "SYSTEM" {
		user.Name = "SYSTEM"
		user.Role = "SYSTEM"
		return nil
	}
	if user.Name == "" && user.InternalUserID != "" {
		databaseUser, err := dbQueries.GetUserById(ctx, user.InternalUserID)
		if err != nil {
			fmt.Println("error occurred during loadAdditionalUserData: ", err)
			return err
		}
		user.Name = databaseUser.Name
	}
	return nil
}
