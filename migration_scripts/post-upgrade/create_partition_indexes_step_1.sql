-- Copyright (c) 2017-2020 VMware, Inc. or its affiliates
-- SPDX-License-Identifier: Apache-2.0

-- generates SQL statement to create non-unique indexes on child partition tables

WITH leaf_partitions (relid) AS
(
   SELECT DISTINCT
      parchildrelid
   FROM
      pg_partition_rule
)
,
leaf_constraints AS
(
   SELECT
      conname,
      c.relname conrel,
      n.nspname relschema,
      cc.relname rel
   FROM
      pg_constraint con
      JOIN
         pg_depend dep
         ON (refclassid, classid, objsubid) =
         (
            'pg_constraint'::regclass,
            'pg_class'::regclass,
            0
         )
         AND refobjid = con.oid
         AND deptype = 'i'
         AND contype IN
         (
            'u',
            'p'
         )
      JOIN
         pg_class c
         ON objid = c.oid
         AND relkind = 'i'
      JOIN
         leaf_partitions
         ON con.conrelid = leaf_partitions.relid
      JOIN
         pg_class cc
         ON cc.oid = con.conrelid
      JOIN
         pg_namespace n
         ON (n.oid = cc.relnamespace)
)
,
indexes AS
(
   SELECT
      n.nspname AS schemaname,
      c.relname AS tablename,
      i.relname AS indexname,
      t.spcname AS tablespace,
      pg_get_indexdef(i.oid) AS indexdef
   FROM
      pg_index x
      JOIN
         leaf_partitions lp
         on lp.relid = x.indrelid
      JOIN
         pg_class c
         ON c.oid = x.indrelid
      JOIN
         pg_class i
         ON i.oid = x.indexrelid
      LEFT JOIN
         pg_namespace n
         ON n.oid = c.relnamespace
      LEFT JOIN
         pg_tablespace t
         ON t.oid = i.reltablespace
   WHERE
      c.relkind = 'r'::"char"
      AND i.relkind = 'i'::"char"
      AND c.relhassubclass = 'f'
      AND x.indisunique = 'f'
)
SELECT
$$SET SEARCH_PATH=$$ || schemaname || $$; $$ || indexdef || $$ ;$$
FROM
   indexes
WHERE
   (
      indexname,
      schemaname,
      tablename
   )
   NOT IN
   (
      SELECT
         conrel,
         relschema,
         rel
      FROM
         leaf_constraints
   )
;
