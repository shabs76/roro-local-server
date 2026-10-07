-- vehicles_talling_history.number_of_keys was TINYINT while vehicles_talling has
-- INT: archiving a tally with more than 127 keys would fail. MODIFY is idempotent.
ALTER TABLE vehicles_talling_history MODIFY number_of_keys INT(4) NOT NULL DEFAULT 0;
