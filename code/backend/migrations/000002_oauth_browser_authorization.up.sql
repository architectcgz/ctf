CREATE TABLE public.oauth_clients (
    id bigserial PRIMARY KEY,
    client_id text NOT NULL UNIQUE,
    client_name text NOT NULL,
    client_uri text,
    redirect_uris jsonb NOT NULL,
    grant_types jsonb NOT NULL,
    response_types jsonb NOT NULL,
    scope text NOT NULL,
    token_endpoint_auth_method text DEFAULT 'none'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.oauth_consents (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    client_id text NOT NULL REFERENCES public.oauth_clients(client_id) ON DELETE CASCADE,
    scope text NOT NULL,
    granted_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone,
    UNIQUE (user_id, client_id, scope)
);

CREATE INDEX idx_oauth_consents_user_client ON public.oauth_consents(user_id, client_id, revoked_at);
