-- Drop table if exists
-- DROP TABLE IF EXISTS landing_page_configs;

CREATE TABLE landing_page_configs (
    id uuid NOT NULL,

    -- Event Info
    event_name varchar NOT NULL,
    event_subtitle varchar NULL,       
    event_creator varchar NULL,
    event_date varchar NOT NULL,
    event_time_start varchar NOT NULL,
    event_time_end varchar NOT NULL,
    event_location varchar NULL,
    logo_image varchar NULL,            

    -- Hero Section
    hero_id varchar NULL,              
    banner_image varchar NULL,          
    banner_color varchar NULL,         
    hero_button_id varchar NULL,       
    hero_button_text varchar NULL,     
    hero_button_link varchar NULL,     

    -- Countdown Section
    hero_countdown_id varchar NULL,        
    hero_countdown_date varchar NULL,      
    hero_countdown_time_start varchar NULL,
    hero_countdown_time_end varchar NULL,  
    hero_countdown_after_text varchar NULL,

    -- Venue Section
    venue_id varchar NULL,             
    venue_image varchar NULL,           
    venue_layout varchar NULL,         
    venue_address varchar NULL,        
    venue_map_link varchar NULL, 
    venue_google varchar NULL,      
     

    -- JSON Data
    terms_and_conditions jsonb NOT NULL DEFAULT '[]'::jsonb,
    faqs jsonb NOT NULL DEFAULT '[]'::jsonb, 

    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT landing_page_configs_pkey PRIMARY KEY (id)
);

-- Indexes for Faster Search
CREATE INDEX idx_landing_page_configs_deleted ON landing_page_configs (deleted);
CREATE INDEX idx_landing_page_configs_event_name ON landing_page_configs (event_name);
CREATE INDEX idx_landing_page_configs_created_at ON landing_page_configs (created_at);