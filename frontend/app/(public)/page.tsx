import Link from 'next/link';

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Hero Section */}
      <div className="max-w-6xl mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold text-gray-900 mb-6">
            Welcome to OrgMind
          </h1>
          <p className="text-xl text-gray-700 mb-8 max-w-3xl mx-auto">
            Enterprise document processing platform with AI-powered knowledge graph generation.
            Transform your documents into actionable insights.
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              href="/signup"
              className="inline-block bg-indigo-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-indigo-700 transition"
            >
              Get Started
            </Link>
            <Link
              href="/signin"
              className="inline-block bg-white text-indigo-600 px-8 py-3 rounded-lg font-semibold border-2 border-indigo-600 hover:bg-indigo-50 transition"
            >
              Sign In
            </Link>
          </div>
        </div>

        {/* Features Section */}
        <div className="grid md:grid-cols-3 gap-8 mb-16">
          <div className="bg-white rounded-lg p-6 shadow-md">
            <div className="text-4xl mb-4">📝</div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              Multiple Input Sources
            </h3>
            <p className="text-gray-600">
              Upload documents, use our rich text editor, or integrate via API.
              Support for various file formats.
            </p>
          </div>

          <div className="bg-white rounded-lg p-6 shadow-md">
            <div className="text-4xl mb-4">🧠</div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              AI-Powered Processing
            </h3>
            <p className="text-gray-600">
              Automatically extract entities and relationships from your documents
              using advanced AI technology.
            </p>
          </div>

          <div className="bg-white rounded-lg p-6 shadow-md">
            <div className="text-4xl mb-4">🔗</div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              Knowledge Graphs
            </h3>
            <p className="text-gray-600">
              Visualize connections between concepts with interactive knowledge graphs
              powered by Zep Cloud.
            </p>
          </div>
        </div>

        {/* Authentication Options */}
        <div className="bg-white rounded-lg p-8 shadow-md max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold text-gray-900 mb-4 text-center">
            Enterprise-Ready Authentication
          </h2>
          <p className="text-gray-600 text-center mb-6">
            Sign in with your preferred method: email/password, Google, Okta, or Office 365
          </p>
          <div className="flex justify-center gap-4">
            <div className="text-3xl">🔐</div>
            <div className="text-3xl">🔍</div>
            <div className="text-3xl">📧</div>
          </div>
        </div>
      </div>
    </div>
  );
}
