INSERT INTO "user" (
    id,
    name,
    email,
    password_hash,
    role,
    deleted,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'Administrator',
    'admin@example.com',
    '$2y$10$examplehashedpasswordadmin',
    'ADMIN',
    false,
    NOW(),
    NOW()
);
