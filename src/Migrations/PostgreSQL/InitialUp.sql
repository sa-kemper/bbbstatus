CREATE TABLE meeting_events
(
    event_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    event_type          VARCHAR(50)  NOT NULL,
    event_timestamp     TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id)
);

CREATE TABLE user_events
(
    event_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    event_type          VARCHAR(50)  NOT NULL,
    event_timestamp     TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);
CREATE TABLE chat_messages
(
    message_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    chat_id             VARCHAR(50)  NOT NULL,
    message_content     TEXT         NOT NULL,
    send_time           TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);


CREATE TABLE polls
(
    poll_id             VARCHAR(255) PRIMARY KEY,
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    question            TEXT         NOT NULL,
    answers             JSONB        NOT NULL,
    created_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP UNIQUE,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);


CREATE TABLE poll_responses
(
    response_id      UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    poll_id          VARCHAR(255) NOT NULL,
    internal_user_id VARCHAR(255) NOT NULL,
    answer_ids       JSONB        NOT NULL,
    response_time    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (poll_id) REFERENCES polls (poll_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);

CREATE INDEX idx_poll_responses_poll_id ON poll_responses (poll_id);
CREATE INDEX idx_user_events_internal_meeting_id ON user_events (internal_meeting_id);
CREATE INDEX idx_chat_messages_internal_meeting_id ON chat_messages (internal_meeting_id);
CREATE INDEX idx_polls_internal_meeting_id ON polls (internal_meeting_id);
CREATE INDEX idx_meetings_external_id ON meetings (external_meeting_id);
CREATE INDEX idx_chat_messages_meeting_time ON chat_messages (internal_meeting_id);
CREATE INDEX idx_polls_meeting ON polls (internal_meeting_id);

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

-- Common event types enum
CREATE TYPE meeting_event_type AS ENUM (
    'meeting-created',
    'meeting-ended'
    );

CREATE TYPE user_event_type AS ENUM (
    'user-joined',
    'user-left',
    'user-presenter-as-signed',
    'user-audio-voice-enabled',
    'user-audio-voice-disabled',
    'user-audio-muted',
    'user-audio-unmuted',
    'user-cam-broadcast-start',
    'user-cam-broadcast-end',
    'user-emoji-changed',
    'user-raise-hand-changed'
    );
