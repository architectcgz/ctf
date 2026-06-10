DO $$
DECLARE
  service_record record;
  parsed_runtime_config jsonb;
BEGIN
  FOR service_record IN
    SELECT id, runtime_config
    FROM public.contest_awd_services
    WHERE runtime_config IS NOT NULL
      AND btrim(runtime_config) <> ''
      AND runtime_config LIKE '%challenge_id%'
  LOOP
    BEGIN
      parsed_runtime_config := service_record.runtime_config::jsonb;
    EXCEPTION WHEN invalid_text_representation THEN
      CONTINUE;
    END;

    IF jsonb_typeof(parsed_runtime_config) = 'object'
      AND parsed_runtime_config ? 'challenge_id' THEN
      UPDATE public.contest_awd_services
      SET runtime_config = (parsed_runtime_config - 'challenge_id')::text
      WHERE id = service_record.id;
    END IF;
  END LOOP;
END $$;
