-- Clickhouse queue wrapper
CREATE TABLE demo_events_queue ON CLUSTER '{cluster}' (
   user_ts String,
   id UInt64,
   message String
) ENGINE = Kafka SETTINGS
    kafka_broker_list = 'KAFKA_HOST:9091',
    kafka_topic_list = 'clickhouse-events',
    kafka_group_name = 'uniq_group_id',
    kafka_format = 'JSONEachRow';

-- Table to store data
CREATE TABLE demo_events_table ON CLUSTER '{cluster}' (
                          topic String,
                          offset UInt64,
                          partition UInt64,
                          timestamp DateTime64,
                          user_ts DateTime64,
                          id UInt64,
                          message String
) Engine = ReplicatedMergeTree('/clickhouse/tables/{shard}/{database}/demo_events_table', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (topic, partition, offset);

-- Delivery pipeline
CREATE MATERIALIZED VIEW readings_queue_mv TO demo_events_table AS
SELECT
        _topic as topic,
        _offset as offset,
        _partition as partition,
        _timestamp as timestamp,                                             -- kafka engine virtual column
        toDateTime64(parseDateTimeBestEffort(user_ts), 6, 'UTC') as user_ts, -- example of complex date parsing
        id, message
FROM demo_events_queue;
