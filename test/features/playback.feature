Feature: Playback

  Background: The app is running
    Given the VOD API is running
    Given Studio API server is running at "localhost:3000"

  Scenario: Master playlist requests
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the body matches file "responses/hls/dbe3q3g6q2kia036/index.m3u8"
    And my "successful" playback request metrics get recorded

  Scenario: Rendition playlist requests
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/720p0/index.m3u8?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the body matches file "responses/hls/dbe3q3g6q2kia036/720p0/index.m3u8"
    And my "successful" playback request metrics get recorded

  Scenario: Gate API deny
    Given the gate API will deny playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the body matches file "responses/hls/unauthorised.m3u8"
    When I query the "/asset/hls/dbe3q3g6q2kia036/foo.ts?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "401"
    And the body matches file "responses/unauthorised"
    And my "failed" playback request metrics get recorded

  Scenario: No token param
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "400"
    And the gate API call was valid

  Scenario: Gate API caching
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    And I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the gate API will be called 1 times

  Scenario: HEAD requests
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint with "HEAD"
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the headers match
      | key                         | value |
      | accept-ranges               | bytes |
      | Access-Control-Allow-Origin | *     |

  Scenario: OPTIONS requests
    Given the gate API will allow playback
    When I query the "/asset/hls/dbe3q3g6q2kia036/index.m3u8?accessKey=secretlpkey" endpoint with "OPTIONS"
    And receive a response within "3" seconds
    Then I get an HTTP response with code "200"
    And the headers match
      | key                         | value                  |
      | accept-ranges               | bytes                  |
      | allow                       | GET, HEAD, OPTIONS     |
      | content-length              | 0                      |
