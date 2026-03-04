INSERT INTO events (
    id,
    slug,
    name,
    status,
    ticket_prefix_code,
    max_ticket_per_tx,
    deleted,
    data_hash,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'rakit-festival-2026',
    'Rakit Festival 2026',
    'PUBLISHED',
    'RF26',
    4,
    false,
    md5(random()::text),
    NOW(),
    NOW()
)
RETURNING id;