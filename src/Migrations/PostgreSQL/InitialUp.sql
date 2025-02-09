CREATE INDEX idx_poll_responses_poll_id ON poll_responses (poll_id);
CREATE INDEX idx_user_events_internal_meeting_id ON user_events (internal_meeting_id);
CREATE INDEX idx_chat_messages_internal_meeting_id ON chat_messages (internal_meeting_id);
CREATE INDEX idx_polls_internal_meeting_id ON polls (internal_meeting_id);
CREATE INDEX idx_meetings_external_id ON meetings (external_meeting_id);
CREATE INDEX idx_chat_messages_meeting_time ON chat_messages (internal_meeting_id);
CREATE INDEX idx_polls_meeting ON polls (internal_meeting_id);
#