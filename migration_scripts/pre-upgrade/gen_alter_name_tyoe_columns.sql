-- Copyright (c) 2017-2020 VMware, Inc. or its affiliates
-- SPDX-License-Identifier: Apache-2.0

-- TODO from pg_upgrade...add postgres copyright attribution?

-- generate ALTER TABLE ALTER COLUMN commands for tables with name-type attributes
SELECT 'ALTER TABLE ' || c.oid::pg_catalog.regclass || ' ALTER COLUMN ' || pg_catalog.quote_ident(a.attname) || ' TYPE VARCHAR(63);'
FROM	pg_catalog.pg_class c,
        pg_catalog.pg_namespace n,
        pg_catalog.pg_attribute a
WHERE	c.oid = a.attrelid AND
        a.attnum > 1 AND
        NOT a.attisdropped AND
        a.atttypid = 'pg_catalog.name'::pg_catalog.regtype AND
        c.relnamespace = n.oid AND
        -- exclude possible orphaned temp tables
        n.nspname !~ '^pg_temp_' AND
        n.nspname !~ '^pg_toast_temp_' AND
        n.nspname NOT IN ('pg_catalog', 'information_schema', 'gp_toolkit')
