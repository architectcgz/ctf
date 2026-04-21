ALTER TABLE contest_awd_services
    ADD COLUMN IF NOT EXISTS service_snapshot TEXT NOT NULL DEFAULT '{}';

UPDATE contest_awd_services AS cas
SET service_snapshot = jsonb_strip_nulls(
    jsonb_build_object(
        'name', COALESCE(NULLIF(src.display_name, ''), src.template_name, src.challenge_title, ''),
        'category', COALESCE(NULLIF(src.template_category, ''), src.challenge_category, ''),
        'difficulty', COALESCE(NULLIF(src.template_difficulty, ''), src.challenge_difficulty, ''),
        'description', COALESCE(src.template_description, ''),
        'service_type', COALESCE(src.template_service_type, ''),
        'deployment_mode', COALESCE(src.template_deployment_mode, ''),
        'flag_mode', COALESCE(src.template_flag_mode, ''),
        'flag_config', jsonb_strip_nulls(
            COALESCE(NULLIF(src.template_flag_config, '')::jsonb, '{}'::jsonb) ||
            jsonb_build_object(
                'flag_type', NULLIF(src.challenge_flag_type, ''),
                'flag_prefix', NULLIF(src.challenge_flag_prefix, '')
            )
        ),
        'defense_entry_mode', COALESCE(src.template_defense_entry_mode, ''),
        'access_config', COALESCE(NULLIF(src.template_access_config, '')::jsonb, '{}'::jsonb),
        'runtime_config', jsonb_strip_nulls(
            COALESCE(NULLIF(src.template_runtime_config, '')::jsonb, '{}'::jsonb) ||
            jsonb_build_object(
                'image_id', CASE WHEN src.challenge_image_id > 0 THEN src.challenge_image_id ELSE NULL END,
                'instance_sharing', NULLIF(src.challenge_instance_sharing, ''),
                'topology', CASE
                    WHEN src.topology_challenge_id IS NOT NULL THEN jsonb_build_object(
                        'entry_node_key', src.topology_entry_node_key,
                        'spec', src.topology_spec
                    )
                    ELSE NULL
                END
            )
        )
    )
)::text
FROM (
    SELECT
        cas_inner.id,
        cas_inner.display_name,
        tpl.name AS template_name,
        tpl.category AS template_category,
        tpl.difficulty AS template_difficulty,
        tpl.description AS template_description,
        tpl.service_type AS template_service_type,
        tpl.deployment_mode AS template_deployment_mode,
        tpl.flag_mode AS template_flag_mode,
        tpl.flag_config AS template_flag_config,
        tpl.defense_entry_mode AS template_defense_entry_mode,
        tpl.access_config AS template_access_config,
        tpl.runtime_config AS template_runtime_config,
        ch.title AS challenge_title,
        ch.category AS challenge_category,
        ch.difficulty AS challenge_difficulty,
        ch.flag_type AS challenge_flag_type,
        ch.flag_prefix AS challenge_flag_prefix,
        ch.image_id AS challenge_image_id,
        ch.instance_sharing AS challenge_instance_sharing,
        topo.challenge_id AS topology_challenge_id,
        topo.entry_node_key AS topology_entry_node_key,
        topo.spec AS topology_spec
    FROM contest_awd_services AS cas_inner
    JOIN awd_service_templates AS tpl
        ON tpl.id = cas_inner.template_id
    LEFT JOIN challenges AS ch
        ON ch.id = cas_inner.challenge_id
    LEFT JOIN challenge_topologies AS topo
        ON topo.challenge_id = ch.id
       AND topo.deleted_at IS NULL
    WHERE cas_inner.deleted_at IS NULL
) AS src
WHERE cas.id = src.id
  AND (
      cas.service_snapshot IS NULL
      OR cas.service_snapshot = ''
      OR cas.service_snapshot = '{}'
  );
