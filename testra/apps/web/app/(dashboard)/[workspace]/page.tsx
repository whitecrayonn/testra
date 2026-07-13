export default function WorkspacePage({
  params,
}: {
  params: { workspace: string };
}) {
  return (
    <main>
      <h1>Workspace {params.workspace}</h1>
    </main>
  );
}
