\c postgres
-- The below SQL alters, where possible, the name data type to varchar(63)
-- in columns other than the first.  For a partition table, SQL is only generated
-- for root partitions as that cascades to the child partitions. Where not possible,
-- a message is logged, and such tables must be manually modified to remove the name
-- data type prior to running gpugprade.

\set VERBOSITY terse

\unset ECHO
CREATE OR REPLACE FUNCTION pg_temp.notsupported(text) RETURNS VOID AS $$
BEGIN
    RAISE WARNING '---------------------------------------------------------------------------------';
    RAISE WARNING 'Removing the name datatype column failed on table ''%''.  You must resolve it manually.',$1;
    RAISE WARNING '---------------------------------------------------------------------------------';
END
$$ LANGUAGE plpgsql;
\set ECHO queries


DO $$ BEGIN ALTER TABLE partition_table_partitioned_by_name_type ALTER COLUMN b TYPE VARCHAR(63); EXCEPTION WHEN feature_not_supported THEN PERFORM pg_temp.notsupported('partition_table_partitioned_by_name_type'); END $$;
DO $$ BEGIN ALTER TABLE table_distributed_by_name_type ALTER COLUMN b TYPE VARCHAR(63); EXCEPTION WHEN feature_not_supported THEN PERFORM pg_temp.notsupported('table_distributed_by_name_type'); END $$;
