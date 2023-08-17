/*
UUID functions from:
https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74

Subject to the following license:

Copyright 2023 Kyle Hubert <kjmph@users.noreply.github.com> (https://github.com/kjmph)
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

CREATE FUNCTION public.uuid_generate_v7()
RETURNS uuid
AS $$
DECLARE
  unix_ts_ms bytea;
  uuid_bytes bytea;
BEGIN
  unix_ts_ms = substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3);
  -- use random v4 uuid as starting point (which has the same variant we need)
  uuid_bytes = uuid_send(gen_random_uuid());
  -- overlay timestamp
  uuid_bytes = overlay(uuid_bytes placing unix_ts_ms from 1 for 6);
  -- set version 7
  uuid_bytes = set_byte(uuid_bytes, 6, (b'0111' || get_byte(uuid_bytes, 6)::bit(4))::bit(8)::int);
  return encode(uuid_bytes, 'hex')::uuid;
END
$$
LANGUAGE plpgsql
VOLATILE;

CREATE FUNCTION public.timestamp_from_uuid_v7(_uuid uuid)
RETURNS timestamp without time zone
LANGUAGE sql
IMMUTABLE PARALLEL SAFE STRICT LEAKPROOF
AS $$
  SELECT to_timestamp(('x0000' || substr(_uuid::text, 1, 8) || substr(_uuid::text, 10, 4))::bit(64)::bigint::numeric / 1000);
$$
;

/*
Now to the actual tables:
*/

CREATE TABLE tuples (
    uuid UUID NOT NULL DEFAULT public.uuid_generate_v7(),
    object TEXT NOT NULL,
    relation TEXT NOT NULL,
    user_ TEXT NOT NULL,
    is_userset BOOLEAN NOT NULL,
    PRIMARY KEY (object, relation, user_)
);

CREATE INDEX idx_tuples_partial_user ON tuples (object, relation, user_, is_userset) WHERE is_userset = FALSE;
CREATE INDEX idx_tuples_partial_userset ON tuples (object, relation, user_, is_userset) WHERE is_userset = TRUE;
CREATE UNIQUE INDEX idx_tuples_uuid ON tuples (uuid);
