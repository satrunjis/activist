import { EntityCreateSurface } from "../../shared/ui/surfaces";
import { PositionCreateForm } from "./PositionCreateForm";

type PositionCreatePanelProps = {
  divisionId: string;
  onSuccess: () => void;
  onClose: () => void;
};

export function PositionCreatePanel({ divisionId, onSuccess, onClose }: PositionCreatePanelProps) {
  return (
    <EntityCreateSurface showTitle={false} onClose={onClose}>
      <PositionCreateForm divisionId={divisionId} onSuccess={onSuccess} onCancel={onClose} />
    </EntityCreateSurface>
  );
}
