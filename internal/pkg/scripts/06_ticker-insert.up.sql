INSERT INTO tickets (
    id,
    "type",
    title,
    status,
    description,
    price,
    total,
    available_qty,
    booked_qty,
    sold_qty,
    is_presale,
    order_priority,
    deleted,
    data_hash,
    created_at
) VALUES
(
    gen_random_uuid(),
    'SILVER',
    'Silver Ticket',
    'AVAILABLE',
    'Access to festival area with standard facilities',
    150000.00,
    500,
    500,   -- available_qty
    0,     -- booked_qty
    0,     -- sold_qty
    false,
    1,
    false,
    md5(random()::text),
    NOW()
),
(
    gen_random_uuid(),
    'GOLD',
    'Gold Ticket',
    'AVAILABLE',
    'Access to festival area with premium seating and fast track entry',
    250000.00,
    300,
    300,
    0,
    0,
    false,
    2,
    false,
    md5(random()::text),
    NOW()
),
(
    gen_random_uuid(),
    'FESTIVAL',
    'Festival Pass',
    'AVAILABLE',
    'Full access to all festival areas and VIP lounge',
    500000.00,
    100,
    100,
    0,
    0,
    false,
    3,
    false,
    md5(random()::text),
    NOW()
);