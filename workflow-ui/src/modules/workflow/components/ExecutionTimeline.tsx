import { Box, Typography, Paper, Chip } from "@mui/material";
import { Timeline, TimelineItem, TimelineSeparator, TimelineConnector, TimelineContent, TimelineDot, TimelineOppositeContent } from "@mui/lab";
import { CheckCircle, Error, PlayArrow, Schedule, Refresh } from "@mui/icons-material";
import { ExecutionTimeline as ExecutionTimelineType, ExecutionTimelineEvent } from "@/types/workflow";

const formatTimestamp = (timestamp: string): string => {
  const date = new Date(timestamp);
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const seconds = date.getSeconds().toString().padStart(2, "0");
  const milliseconds = date.getMilliseconds().toString().padStart(3, "0");
  return `${hours}:${minutes}:${seconds}.${milliseconds}`;
};

interface ExecutionTimelineProps {
  timeline: ExecutionTimelineType;
}

const getEventIcon = (type: string) => {
  if (type.includes("STARTED")) {
    return <PlayArrow />;
  }
  if (type.includes("SUCCEEDED")) {
    return <CheckCircle />;
  }
  if (type.includes("FAILED")) {
    return <Error />;
  }
  if (type.includes("SKIPPED")) {
    return <Schedule />;
  }
  if (type.includes("RETRY")) {
    return <Refresh />;
  }
  return <Schedule />;
};

const getEventColor = (type: string): "primary" | "success" | "error" | "warning" | "info" => {
  if (type.includes("SUCCEEDED")) {
    return "success";
  }
  if (type.includes("FAILED")) {
    return "error";
  }
  if (type.includes("SKIPPED")) {
    return "warning";
  }
  if (type.includes("STARTED")) {
    return "primary";
  }
  if (type.includes("RETRY")) {
    return "warning";
  }
  return "info";
};

const getEventLabel = (event: ExecutionTimelineEvent): string => {
  if (event.message) {
    return event.message;
  }
  
  if (event.type === "EXECUTION_STARTED") {
    return "Execution started";
  }
  if (event.type === "EXECUTION_SUCCEEDED") {
    return "Execution completed successfully";
  }
  if (event.type === "EXECUTION_FAILED") {
    return "Execution failed";
  }
  if (event.type === "NODE_STARTED") {
    return `Node "${event.nodeId}" started`;
  }
  if (event.type === "NODE_SUCCEEDED") {
    return `Node "${event.nodeId}" completed`;
  }
  if (event.type === "NODE_FAILED") {
    return `Node "${event.nodeId}" failed`;
  }
  if (event.type === "NODE_SKIPPED") {
    return `Node "${event.nodeId}" skipped`;
  }
  if (event.type === "NODE_RETRY") {
    return `Node "${event.nodeId}" retrying`;
  }
  
  return event.type;
};

const formatDuration = (ms?: number): string => {
  if (!ms) return "";
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
  return `${(ms / 60000).toFixed(2)}m`;
};

export const ExecutionTimeline = ({ timeline }: ExecutionTimelineProps) => {
  const sortedEvents = [...timeline.timeline].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        Execution Timeline
      </Typography>
      
      <Box sx={{ mt: 2 }}>
        <Timeline>
          {sortedEvents.map((event, index) => (
            <TimelineItem key={index}>
              <TimelineOppositeContent sx={{ flex: 0.2, pt: 2 }}>
                <Typography variant="caption" color="text.secondary">
                  {formatTimestamp(event.timestamp)}
                </Typography>
                {event.durationMs && (
                  <Typography variant="caption" display="block" color="text.secondary">
                    {formatDuration(event.durationMs)}
                  </Typography>
                )}
              </TimelineOppositeContent>
              <TimelineSeparator>
                <TimelineDot color={getEventColor(event.type)}>
                  {getEventIcon(event.type)}
                </TimelineDot>
                {index < sortedEvents.length - 1 && <TimelineConnector />}
              </TimelineSeparator>
              <TimelineContent sx={{ pt: 2 }}>
                <Box>
                  <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 0.5 }}>
                    <Chip
                      label={event.type.replace(/_/g, " ")}
                      size="small"
                      color={getEventColor(event.type)}
                      variant="outlined"
                    />
                    {event.nodeId && (
                      <Chip label={`Node: ${event.nodeId}`} size="small" variant="outlined" />
                    )}
                    {event.executor && (
                      <Chip label={event.executor} size="small" variant="outlined" />
                    )}
                  </Box>
                  <Typography variant="body2" sx={{ mb: 1 }}>
                    {getEventLabel(event)}
                  </Typography>
                  {event.payload && Object.keys(event.payload).length > 0 && (
                    <Box
                      sx={{
                        mt: 1,
                        p: 1,
                        bgcolor: "grey.50",
                        borderRadius: 1,
                        maxHeight: 150,
                        overflow: "auto",
                      }}
                    >
                      <Typography variant="caption" component="pre" sx={{ fontSize: "0.75rem" }}>
                        {JSON.stringify(event.payload, null, 2)}
                      </Typography>
                    </Box>
                  )}
                </Box>
              </TimelineContent>
            </TimelineItem>
          ))}
        </Timeline>
      </Box>
      
      {sortedEvents.length === 0 && (
        <Typography variant="body2" color="text.secondary" sx={{ textAlign: "center", py: 4 }}>
          No timeline events available
        </Typography>
      )}
    </Paper>
  );
};

