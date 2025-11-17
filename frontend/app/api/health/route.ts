// Health check endpoint for Cloud Run and monitoring
export async function GET() {
  return Response.json(
    {
      status: 'healthy',
      service: 'orgmind-frontend',
      timestamp: new Date().toISOString(),
    },
    { status: 200 }
  );
}
