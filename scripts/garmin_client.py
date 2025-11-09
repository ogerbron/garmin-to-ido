#!/usr/bin/env python3
"""
Python wrapper for Garmin Connect API using garminconnect library.
This script can be called from Go to handle authentication and data fetching.
"""

import sys
import json
import argparse
from datetime import datetime, timedelta
from garminconnect import Garmin


def login(username, password):
    """Login to Garmin Connect and return client."""
    try:
        client = Garmin(username, password)
        client.login()
        return client
    except Exception as e:
        print(f"Login failed: {e}", file=sys.stderr)
        sys.exit(1)


def get_activities(client, date_str):
    """Get activities for a specific date."""
    try:
        # Parse date
        date = datetime.strptime(date_str, "%Y-%m-%d")

        # Get activities for the date range
        activities = client.get_activities_by_date(
            date.strftime("%Y-%m-%d"),
            (date + timedelta(days=1)).strftime("%Y-%m-%d")
        )

        return activities
    except Exception as e:
        print(f"Failed to get activities: {e}", file=sys.stderr)
        sys.exit(1)


def download_activity(client, activity_id, format="FIT"):
    """Download activity in specified format (FIT, GPX, TCX, etc.)."""
    try:
        if format.upper() == "FIT":
            data = client.download_activity(activity_id, dl_fmt=client.ActivityDownloadFormat.ORIGINAL)
        elif format.upper() == "GPX":
            data = client.download_activity(activity_id, dl_fmt=client.ActivityDownloadFormat.GPX)
        elif format.upper() == "TCX":
            data = client.download_activity(activity_id, dl_fmt=client.ActivityDownloadFormat.TCX)
        else:
            # Default to ORIGINAL which is usually FIT
            data = client.download_activity(activity_id, dl_fmt=client.ActivityDownloadFormat.ORIGINAL)
        return data
    except Exception as e:
        print(f"Failed to download activity: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Garmin Connect API wrapper")
    parser.add_argument("--username", required=True, help="Garmin username")
    parser.add_argument("--password", required=True, help="Garmin password")
    parser.add_argument("--command", required=True, choices=["get-activities", "download-activity"],
                        help="Command to execute")
    parser.add_argument("--date", help="Date for get-activities (YYYY-MM-DD)")
    parser.add_argument("--activity-id", type=int, help="Activity ID for download-activity")
    parser.add_argument("--output", help="Output file for download-activity")
    parser.add_argument("--format", default="FIT", choices=["FIT", "GPX", "TCX"],
                        help="Download format for activity (default: FIT)")

    args = parser.parse_args()

    # Login
    client = login(args.username, args.password)

    if args.command == "get-activities":
        if not args.date:
            print("--date is required for get-activities", file=sys.stderr)
            sys.exit(1)

        activities = get_activities(client, args.date)

        # Filter for cycling activities and output as JSON
        bike_activities = []
        for activity in activities:
            activity_type = activity.get("activityType", {}).get("typeKey", "").lower()
            if "cycling" in activity_type or "bike" in activity_type or "biking" in activity_type:
                # Parse and reformat startTimeLocal to ISO8601
                start_time = activity.get("startTimeLocal")
                if start_time and " " in start_time:
                    # Convert "YYYY-MM-DD HH:MM:SS" to "YYYY-MM-DDTHH:MM:SSZ"
                    start_time = start_time.replace(" ", "T") + "Z"

                bike_activities.append({
                    "activityId": activity.get("activityId"),
                    "activityName": activity.get("activityName"),
                    "activityType": activity.get("activityType", {}).get("typeKey"),
                    "startTimeLocal": start_time,
                    "distance": activity.get("distance", 0),
                    "duration": activity.get("duration", 0),
                    "averageSpeed": activity.get("averageSpeed", 0),
                    "calories": activity.get("calories", 0)
                })

        print(json.dumps(bike_activities, indent=2))

    elif args.command == "download-activity":
        if not args.activity_id:
            print("--activity-id is required for download-activity", file=sys.stderr)
            sys.exit(1)

        activity_data = download_activity(client, args.activity_id, format=args.format)

        if args.output:
            with open(args.output, "wb") as f:
                f.write(activity_data)
            print(f"Activity downloaded to {args.output}")
        else:
            # Output to stdout
            sys.stdout.buffer.write(activity_data)


if __name__ == "__main__":
    main()
