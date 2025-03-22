package tracks

/*
steps:
1. if not found, use the below query to retrieve all songs by artists that are related to the submitted song
        SELECT s.id,
               s.name,
               s.length,
               s.tempo,
               s.album,
               s.release_date
        FROM songs s
                 JOIN artists_songs "as" on s.id = "as".song_id
        WHERE "as".artist_id = ANY (SELECT as2.artist_id
                                    FROM artists_songs as2
                                    WHERE as2.song_id = '{song_id}')
        GROUP BY s.id;

2. if 0 records, then it is a new song which ok. if 1 record, then the same song has been added before, but is not a duplicate, which is ok.
3. if 2 or more records, then follow these steps:
    1. filter down potential duplicates based on length (within 0.17 min) and tempo (within 5 bpm). if no records, return
    2. detect if the exiting song is a remix of the submitted song (based on the word 'remix' in the song name) and filter records accordingly
    3. out of the remaining results, filter down based on the song name (fuzzy search - original impl was in python
       using thefuzz library with a threshold of 90). if no records, return
4. at this point, any remaining records are the same song. pick the version that has more existing tracks from the album in the database
*/
