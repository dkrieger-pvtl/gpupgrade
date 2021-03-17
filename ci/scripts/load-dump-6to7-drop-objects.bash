#!/bin/bash
#
# Copyright (c) 2017-2021 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

set -eux -o pipefail


export GPHOME_SOURCE=/usr/local/greenplum-db-source
export GPHOME_TARGET=/usr/local/greenplum-db-target

export PGPORT=5432

echo "getting all databases"

# TODO: the database ["funny copy"db'\''with\\quotes"] creates issues below, and I don't
# think we need to drop anything from it, so exclude it here
databases=$(ssh -n gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression --tuples-only --no-align --field-separator ' ' <<SQL_EOF
        SELECT datname
        FROM	pg_database
        WHERE	datname != 'template0' and datname not like 'funny%';
SQL_EOF
")

# for initialize to work
echo 'a few tables have WITH OIDS set.....alter the tables to be WITHOUT OIDS'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
      ALTER TABLE public.emp SET WITHOUT OIDS;
      ALTER TABLE public.stud_emp SET WITHOUT OIDS;
      ALTER TABLE public.tabwithoids SET WITHOUT OIDS;
      ALTER TABLE public.tenk1 SET WITHOUT OIDS;
      ALTER TABLE public.tt7 SET WITHOUT OIDS;
      ALTER TABLE qp_dml_oids.dml_ao SET WITHOUT OIDS;
      ALTER TABLE qp_dml_oids.dml_heap_check_r SET WITHOUT OIDS;
      ALTER TABLE qp_dml_oids.dml_heap_p SET WITHOUT OIDS;
      ALTER TABLE qp_dml_oids.dml_heap_r SET WITHOUT OIDS;
      ALTER TABLE qp_dml_oids.dml_heap_with_oids SET WITHOUT OIDS;
      ALTER TABLE sort_schema.gpsort_alltypes SET WITHOUT OIDS;
SQL_EOF
"

# for initialize to work
echo 'unknown data type not supported'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP TABLE public.aocs_unknown;
SQL_EOF
"

# for initialize to work
# drop all functions using plpython2u
# we should do this:
# SELECT DISTINCT proname,probin,pronargs,proargtypes FROM pg_catalog.pg_proc  WHERE prolang = 13 AND probin LIKE '%plpython2' AND  oid >= 16384;
#              select typname from pg_type where oid=<progarytpes> from above;
# That is, get all function names and types to construct the below
echo 'drop all functions using plpythonu/plpython2'
echo "${databases}" | while read -r database; do
    if [[ -n "${database}" ]]; then
        ssh -n gpadmin@mdw "
        set -x
        source /usr/local/greenplum-db-source/greenplum_path.sh
        export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

        psql -d ${database} <<SQL_EOF
            DROP FUNCTION IF EXISTS plpython_validator(oid) CASCADE;
            DROP FUNCTION IF EXISTS plpython_inline_handler(internal) CASCADE;
            DROP FUNCTION IF EXISTS plpython_call_handler() CASCADE;
SQL_EOF
"
    fi
done

# gpupgrade execute
# the big one...drop all partition tables until pg_dump is fixed...
echo 'drop all partition tables'
echo "${databases}" | while read -r database; do
    if [[ -n "${database}" ]]; then
        root_partitions=$(ssh -n gpadmin@mdw "
            set -x

            source /usr/local/greenplum-db-source/greenplum_path.sh
            export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

            psql -d ${database}  --tuples-only --no-align --field-separator ' ' <<SQL_EOF
                    SELECT DISTINCT schemaname, tablename FROM pg_partitions;
SQL_EOF
        ")

        echo "${root_partitions}" | while read -r root_schema root_table; do
            if [[ -n "${root_table}" ]]; then
                ssh -n gpadmin@mdw "
                    set -x
                    source /usr/local/greenplum-db-source/greenplum_path.sh
                    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

                    psql -d ${database} << SQL_EOF
                        DROP TABLE IF EXISTS ${root_schema}.${root_table} CASCADE;
SQL_EOF
                "
            fi
        done
    fi
done

# gpugprade execute: pg_partition is not in 7X....drop anything using it
echo 'drop all views using obsolete pg_partition/pg_partition_rule'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP VIEW IF EXISTS mpp7164.partagain CASCADE;
        DROP VIEW IF EXISTS mpp7164.partlist CASCADE;
        DROP VIEW IF EXISTS mpp7164.partrank CASCADE;
        DROP VIEW IF EXISTS public.redundantly_named_part;
SQL_EOF
"

# gpugprade execute:  type "abstime" does not exist
echo 'drop all objects using abstime'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS public.all_legacy_types;
SQL_EOF
"

# gpugprade execute: gp_default_storage_options
echo 'drop all tables using gp_default_storage_options'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP DATABASE dsp1;
        DROP DATABASE dsp2;
        DROP DATABASE dsp3;
SQL_EOF
"

# gpugprade execute: drop all operations named '=>'
echo 'drop operator =>'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP OPERATOR public.=>(bigint,NONE);
SQL_EOF
"

# gpugprade execute: drop MATERIALIZED VIEWS
echo 'drop materialized views'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP MATERIALIZED VIEW m_aocs;
        DROP MATERIALIZED VIEW m_ao;
SQL_EOF
"

# gpugprade execute: drop certain triggers
echo 'drop certain triggers'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP TRIGGER after_ins_stmt_trig ON public.main_table;
SQL_EOF
"

# gpugprade execute: drop external tables
echo 'drop external tables'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d gpfdist_regression <<SQL_EOF
        DROP EXTERNAL TABLE cat_sqlout_result;
        DROP EXTERNAL TABLE create_pipe;
        DROP EXTERNAL TABLE exttab1_gpfdist_status;
        DROP EXTERNAL TABLE gpfdist2_start;
        DROP EXTERNAL TABLE gpfdist2_stop;
        DROP EXTERNAL TABLE gpfdist_status;
        DROP EXTERNAL TABLE pipe_ext1;
        DROP EXTERNAL TABLE pipe_ext2;                                                                                                                                                                            ;
        DROP EXTERNAL TABLE write_pipe;
SQL_EOF

    psql -d mapred_regression <<SQL_EOF
        DROP EXTERNAL TABLE env_master CASCADE;
        DROP EXTERNAL TABLE env_segment CASCADE;
SQL_EOF

    psql -d exttab_db <<SQL_EOF
        DROP EXTERNAL TABLE  exttab_permissions_1 CASCADE;
SQL_EOF

    psql -d isolation2test <<SQL_EOF
        DROP EXTERNAL TABLE IF EXISTS public.ext_delim_off CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_cursor_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_cursor_2 CASCADE;
SQL_EOF
"
# gpugprade execute: drop certain trigger statements
# select pg_get_triggerdef(f.oid) from (select oid from pg_trigger) as f;
# grep that for "FOR EACH STATEMENT" and drop those:
echo 'drop certain triggers'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d regression <<SQL_EOF
        DROP TRIGGER IF EXISTS after_upd_b_stmt_trig on public.main_table;
        DROP TRIGGER IF EXISTS after_upd_stmt_trig on public.main_table;
        DROP TRIGGER IF EXISTS before_stmt_trig on public.main_table;
        DROP TRIGGER IF EXISTS before_ins_stmt_trig on public.main_table;
        DROP TRIGGER IF EXISTS before_upd_a_stmt_trig on public.main_table;
        DROP TRIGGER IF EXISTS foo_as_trigger on  test_expand_table.table_with_update_trigger;
        DROP TRIGGER IF EXISTS foo_bs_trigger on  test_expand_table.table_with_update_trigger;
SQL_EOF
"

# gpugprade execute: drop certain toast tables
# the EXTERNAL drops here are also toast-related and not external table related
echo 'drop certain toast tables'
ssh gpadmin@mdw "
    set -x

    source /usr/local/greenplum-db-source/greenplum_path.sh
    export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

    psql -d contrib_regression <<SQL_EOF
        DROP TABLE IF EXISTS public.a_aoco_table_with_zstd_compression CASCADE;
        DROP TABLE IF EXISTS public.zstd_leak_test CASCADE;
        DROP TABLE IF EXISTS public.zstdtest CASCADE;
SQL_EOF

    psql -d isolation2test <<SQL_EOF
        DROP TABLE IF EXISTS public.a_aoco_table_with_zstd_compression CASCADE;
        DROP TABLE IF EXISTS public.zstd_leak_test CASCADE;
        DROP TABLE IF EXISTS public.zstdtest CASCADE;
        DROP TABLE IF EXISTS public.aocs_rle_upgrade_test CASCADE;
        DROP TABLE IF EXISTS public.aocs_upgrade_test CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS alter_ao_part_exch_column.exh_ao_ao CASCADE;
        DROP TABLE IF EXISTS alter_ao_part_exch_column.exh_ao_co CASCADE;
        DROP TABLE IF EXISTS alter_ao_part_exch_column.exh_ao_heap CASCADE;
        DROP TABLE IF EXISTS alter_ao_part_exch_column.exh_co_ao CASCADE;
        DROP TABLE IF EXISTS alter_ao_part_exch_row.exh_co_ao CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_col_ddl_column.sto_alt_uao1 CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_constraint_column.sto_alt_uao2_constraint CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_index_column.sto_alt_uao3_idx CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_setdefault_column.sto_alt_uao1_default CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_setstorage_column.sto_alt_uao1_setstorage CASCADE;
        DROP TABLE IF EXISTS alter_ao_table_statistics_column.sto_alt_uao2_stats CASCADE;
        DROP TABLE IF EXISTS analyze_ao_table_every_dml_column.sto_uao_city_analyze_everydml CASCADE;
        DROP TABLE IF EXISTS blocksize_column.uao_blocksize_2048k CASCADE;
        DROP TABLE IF EXISTS blocksize_column.uao_blocksize_8k CASCADE;
        DROP TABLE IF EXISTS compresstype_column.uao_tab_compress_none CASCADE;
        DROP TABLE IF EXISTS compresstype_column.uao_tab_compress_zlib1 CASCADE;
        DROP TABLE IF EXISTS compresstype_column.uao_tab_compress_zlib9 CASCADE;
        DROP TABLE IF EXISTS create_ao_table_500cols_column.sto_uao_500cols CASCADE;
        DROP TABLE IF EXISTS create_ao_tables_column.sto_uao_1 CASCADE;
        DROP TABLE IF EXISTS create_ao_tables_column.sto_uao_2 CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS create_ao_tables_column.sto_uao_3 CASCADE;
        DROP TABLE IF EXISTS create_ao_tables_column.sto_uao_8 CASCADE;
        DROP TABLE IF EXISTS create_ao_tables_column.sto_uao_9 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS gpexplain.dummy_ext_tab CASCADE;
        DROP TABLE IF EXISTS public.ao_crtb_with_row_zlib_8192_1_ctas CASCADE;
        DROP TABLE IF EXISTS public.aocs_compress_table CASCADE;
        DROP TABLE IF EXISTS public.aocs_index_cols CASCADE;
        DROP TABLE IF EXISTS public.aocs_with_domain_constraint CASCADE;
        DROP TABLE IF EXISTS public.aocssizetest CASCADE;
        DROP TABLE IF EXISTS public.bulk_rle_tab CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.check_cursor_files CASCADE;
        DROP TABLE IF EXISTS public.ck_ct_co_analyze1 CASCADE;
        DROP TABLE IF EXISTS public.co CASCADE;
        DROP TABLE IF EXISTS public.co_cr_sub_partzlib8192_1_2_defexch CASCADE;
        DROP TABLE IF EXISTS public.co_cr_sub_partzlib8192_1_2_exch CASCADE;
        DROP TABLE IF EXISTS public.co_cr_sub_partzlib8192_1_defexch CASCADE;
        DROP TABLE IF EXISTS public.co_cr_sub_partzlib8192_1_exch CASCADE;
        DROP TABLE IF EXISTS public.co_crtb_with_strg_dir_and_col_ref_1 CASCADE;
        DROP TABLE IF EXISTS public.co_crtb_with_strg_dir_and_col_ref_1_uncompr CASCADE;
        DROP TABLE IF EXISTS public.co_large_and_bulk_content CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS public.co_serial CASCADE;
        DROP TABLE IF EXISTS public.co_t CASCADE;
        DROP TABLE IF EXISTS public.co_wt_sub_partrle_type8192_1_2_defexch CASCADE;
        DROP TABLE IF EXISTS public.co_wt_sub_partrle_type8192_1_2_exch CASCADE;
        DROP TABLE IF EXISTS public.co_wt_sub_partrle_type8192_1_defexch CASCADE;
        DROP TABLE IF EXISTS public.co_wt_sub_partrle_type8192_1_exch CASCADE;
        DROP TABLE IF EXISTS public.col_large_content_block CASCADE;
        DROP TABLE IF EXISTS public.col_large_content_block_add_col CASCADE;
        DROP TABLE IF EXISTS public.decodetimestamptz CASCADE;
        DROP TABLE IF EXISTS public.delta_all CASCADE;
        DROP TABLE IF EXISTS public.delta_alter CASCADE;
        DROP TABLE IF EXISTS public.delta_bitmap_ins CASCADE;
        DROP TABLE IF EXISTS public.delta_btree_ins CASCADE;
        DROP TABLE IF EXISTS public.delta_ins_bitmap CASCADE;
        DROP TABLE IF EXISTS public.delta_ins_btree CASCADE;
        DROP TABLE IF EXISTS public.delta_none CASCADE;
        DROP TABLE IF EXISTS public.delta_zlib CASCADE;
        DROP TABLE IF EXISTS public.dml_co_p CASCADE;
        DROP TABLE IF EXISTS public.dml_co_r CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.echotable CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP EXTERNAL TABLE IF EXISTS public.exttab_subq_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_subq_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_subtxs_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_subtxs_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_txs_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_txs_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.ext_invalid_host CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_3 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_4 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_5 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_6 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_7 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_basic_error_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_constraints_1 CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP EXTERNAL TABLE IF EXISTS public.exttab_cte_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_cte_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_first_reject_limit_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_first_reject_limit_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_heap_join_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_limit_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_limit_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_permissions_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_permissions_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_permissions_3 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_udfs_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_udfs_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_union_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_union_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_views_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_views_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_windows_1 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttab_windows_2 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.exttest CASCADE;
        DROP TABLE IF EXISTS public.heap_can CASCADE;
        DROP TABLE IF EXISTS public.mpp17012_compress_test2 CASCADE;
        DROP TABLE IF EXISTS public.multi_segfile_bitab CASCADE;
        DROP TABLE IF EXISTS public.multi_segfile_tab CASCADE;
        DROP TABLE IF EXISTS public.multi_segfile_toast CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS public.multi_segfile_zlibtab CASCADE;
        DROP TABLE IF EXISTS public.multivarblock_bitab CASCADE;
        DROP TABLE IF EXISTS public.multivarblock_tab CASCADE;
        DROP TABLE IF EXISTS public.multivarblock_toast CASCADE;
        DROP TABLE IF EXISTS public.multivarblock_zlibtab CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.ret_too_many_uris CASCADE;
        DROP TABLE IF EXISTS public.rle_block_boundary CASCADE;
        DROP TABLE IF EXISTS public.subt_reindex_co CASCADE;
        DROP TABLE IF EXISTS public.t_ao CASCADE;
        DROP TABLE IF EXISTS public.t_ao_a CASCADE;
        DROP TABLE IF EXISTS public.t_ao_b CASCADE;
        DROP TABLE IF EXISTS public.t_ao_d CASCADE;
        DROP TABLE IF EXISTS public.t_ao_enc CASCADE;
        DROP TABLE IF EXISTS public.t_ao_enc_a CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.t_ext_r CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.table_env CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.table_qry CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.tableless_ext CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.tbl_ext_gpformatter CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.test2 CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS public.test_table_co_with_toast CASCADE;
        DROP TABLE IF EXISTS public.trigger_aocs_test CASCADE;
        DROP TABLE IF EXISTS public.uaocs_drop_column_update CASCADE;
        DROP TABLE IF EXISTS public.uaocs_index_stats CASCADE;
        DROP TABLE IF EXISTS public.vfao CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.wet_pos4 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS public.wet_too_many_uris CASCADE;
        DROP TABLE IF EXISTS qf.bar4 CASCADE;
        DROP EXTERNAL TABLE IF EXISTS qp_orca_fallback.ext_table_no_fallback CASCADE;
        DROP TABLE IF EXISTS stat_co1.stat_co_t1 CASCADE;
        DROP TABLE IF EXISTS stat_co2.stat_co_t2 CASCADE;
        DROP TABLE IF EXISTS stat_co3.stat_co_t3 CASCADE;
        DROP TABLE IF EXISTS stat_co4.stat_co_t4 CASCADE;
        DROP TABLE IF EXISTS stat_co5.stat_co_t5 CASCADE;
        DROP TABLE IF EXISTS stat_co6.stat_co_t6 CASCADE;
        DROP TABLE IF EXISTS stat_co7.stat_co_t7 CASCADE;
        DROP TABLE IF EXISTS subselect_gp.t3cozlib CASCADE;
        DROP TABLE IF EXISTS uao_allalter_column.uao_allalter CASCADE;
        DROP TABLE IF EXISTS uao_allalter_column.uao_allalter_uncompr CASCADE;
        DROP TABLE IF EXISTS uao_dml_column.mytab_column CASCADE;
SQL_EOF

    psql -d regression <<SQL_EOF
        DROP TABLE IF EXISTS uao_dml_select_column.city_uao CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.city_uao_union CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.city_uao_using CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.city_uao_where CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.country_uao CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.country_uao_subq CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.country_uao_union CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.country_uao_using CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.country_uao_where CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.countrylanguage_uao CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.countrylanguage_uao_union CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.countrylanguage_uao_using CASCADE;
        DROP TABLE IF EXISTS uao_dml_select_column.countrylanguage_uao_where CASCADE;
SQL_EOF
"
